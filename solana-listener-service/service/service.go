package service

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/messages"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
)

type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu        sync.RWMutex
	status    string // started, healthy, nats_failed, replay,
	replaying int32
	startedAt time.Time

	mintEvents    uint64
	accountEvents uint64

	programEventsSeen uint64
	logEventsSeen     uint64

	logsBacklog    *persistence.Backlog
	programBacklog *persistence.Backlog

	seen *expirable.LRU[string, struct{}]

	logIDBase     int
	programIDBase int

	wsUrl         string
	rpcUrl        string
	logSocket     *ownhttp.ManagedWS
	programSocket *ownhttp.ManagedWS

	outFile *os.File

	EventStore   *system.NatsEventStore
	QueueManager *messages.QueueManager
}

func New() *Service {
	wsUrl := helpers.GetEnvOrFail("HELIUS_WS_URL")
	rpcUrl := helpers.GetEnvOrFail("HELIUS_RPC_URL")

	s := &Service{
		wsUrl:         wsUrl,
		rpcUrl:        rpcUrl,
		status:        "started",
		outFile:       nil,
		logIDBase:     rand.Intn(90000) + 10000,
		programIDBase: rand.Intn(90000) + 10000,
	}

	if s.startedAt.IsZero() {
		s.startedAt = time.Now().UTC()
	}

	serviceName := constants.SolanaListenerServiceName
	s.QueueManager = messages.NewManager(serviceName, "support@moonmap.io", "MoonMap Support", s.startedAt)
	s.EventStore = system.NewEventStore(serviceName)

	s.mu.Lock()
	if s.EventStore.GetConn().IsConnected() {
		s.status = "healthy"
	}
	s.mu.Unlock()

	s.EventStore.DisconnectErrHandler = s.DisconnectErrHandler
	s.EventStore.ReconnectHandler = s.ReconnectHandler

	maxEntries, _ := strconv.Atoi(helpers.GetEnv("SEEN_MAX", "20000"))
	ttl, _ := time.ParseDuration(helpers.GetEnv("SEEN_TTL", "10m"))
	lru := expirable.NewLRU[string, struct{}](maxEntries, nil, ttl)
	s.seen = lru

	return s
}

func (s *Service) DisconnectErrHandler(nc *nats.Conn, err error) {
	s.SetStatus("nats_failed")
	msg := "NATS down, switching to JSONL fallback"
	logrus.Warn(msg)
	s.QueueManager.EnqueueWarn(s.QueueManager.ServiceName, msg)
}

// func (s *Service) ReconnectHandler(nc *nats.Conn) {
// 	msg := "NATS back, replaying backlog"
// 	logrus.Info(msg)
// 	s.SetStatus("replay")
// 	s.QueueManager.EnqueueInfo(s.QueueManager.ServiceName, msg)

// 	if atomic.CompareAndSwapInt32(&s.replaying, 0, 1) {
// 		go func() {
// 			defer atomic.StoreInt32(&s.replaying, 0)

// 			s.ReplayFromBacklogs()

// 			// Espera activa breve hasta que no haya pendientes
// 			ticker := time.NewTicker(500 * time.Millisecond)
// 			defer ticker.Stop()

// 			for {
// 				if s.noPendingBacklogs() {
// 					break
// 				}
// 				select {
// 				case <-s.ctx.Done():
// 					return
// 				case <-ticker.C:
// 				}
// 			}

// 			logrus.Info("Backlogs empty after replay; shutting down gracefully")
// 			s.cancel()
// 		}()
// 	}
// }

func (s *Service) ReconnectHandler(nc *nats.Conn) {
	msg := "NATS back, replaying backlog"
	logrus.Info(msg)
	s.SetStatus("replay")
	s.QueueManager.EnqueueInfo(s.QueueManager.ServiceName, msg)

	if atomic.CompareAndSwapInt32(&s.replaying, 0, 1) {
		go func() {
			defer atomic.StoreInt32(&s.replaying, 0)

			s.ReplayFromBacklogs()

			secs := helpers.GetEnvInt("SHUTDOWN_STABLE_SECONDS", 10)
			stableWindow := time.Duration(secs) * time.Second
			tick := time.NewTicker(200 * time.Millisecond)
			defer tick.Stop()

			var sinceClear time.Time

			for {
				if s.ctx.Err() != nil {
					return
				}

				clear := s.noPendingBacklogs() &&
					len(s.logSocket.Messages) == 0 &&
					len(s.programSocket.Messages) == 0

				if clear {
					if sinceClear.IsZero() {
						sinceClear = time.Now()
					}
					if time.Since(sinceClear) >= stableWindow {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						_ = s.EventStore.AwaitAsyncPublishes(ctx)
						cancel()
						logrus.Info("Backlogs empty and queues idle; shutting down after stable window")
						s.cancel()
						return
					}
				} else {
					sinceClear = time.Time{}
				}

				select {
				case <-s.ctx.Done():
					return
				case <-tick.C:
				}
			}
		}()
	}
}

// func (s *Service) ReconnectHandler(nc *nats.Conn) {
// 	msg := "NATS back, replaying backlog"
// 	logrus.Info(msg)
// 	s.SetStatus("replay")
// 	s.QueueManager.EnqueueInfo(s.QueueManager.ServiceName, msg)

// 	if atomic.CompareAndSwapInt32(&s.replaying, 0, 1) {
// 		go func() {
// 			defer atomic.StoreInt32(&s.replaying, 0)
// 			s.ReplayFromBacklogs()
// 		}()
// 	}
// }

func (s *Service) Config(ctx context.Context, cancel context.CancelFunc) {
	s.ctx = ctx
	s.cancel = cancel

	LogSubscribeChannelLength := helpers.GetEnvInt("LOG_SUBSCRIBE_CHANNEL_LENGTH", 10000)
	s.logSocket = &ownhttp.ManagedWS{
		Name:      "LogSubscribe",
		Url:       s.wsUrl,
		OnStatus:  s.OnStatus,
		OnConnect: s.SubscribeLogs,
		Messages:  make(chan []byte, LogSubscribeChannelLength),
	}

	ProgramSubscribeChannelLength := helpers.GetEnvInt("PROGRAM_SUBSCRIBE_CHANNEL_LENGTH", 20000)
	s.programSocket = &ownhttp.ManagedWS{
		Name:      "ProgramSubscribe",
		Url:       s.wsUrl,
		OnStatus:  s.OnStatus,
		OnConnect: s.SubscribeProgram,
		Messages:  make(chan []byte, ProgramSubscribeChannelLength),
	}

	s.QueueManager.SetContext(s.ctx)

	var size100mb int64 = 100 * 1024 * 1024
	s.logsBacklog = persistence.NewBacklog("./data/backlog-logs", size100mb, 5)
	s.programBacklog = persistence.NewBacklog("./data/backlog-program", size100mb, 100)

}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx, sys.GetCancel())

		s.StartMetricsLogger()
		s.QueueManager.StartAggregator(5 * time.Second)
		s.logSocket.Start(s.ctx)
		s.programSocket.Start(s.ctx)

		go s.wsProcessor(s.logSocket.Messages, s.handleLogsMessage)
		go s.wsProcessor(s.programSocket.Messages, s.handleProgramMessage)

		s.QueueManager.EnqueueInfo(s.QueueManager.ServiceName, "Service running")
		ownhttp.NewServer(ctx, constants.SolanaListenerServiceName, sys.Bind, s.routes(), nil)
		<-ctx.Done()
		s.Stop()
	})

}

// func (s *Service) Stop() {
// 	s.logSocket.Close()
// 	s.programSocket.Close()
// 	logrus.Info("Service dependencies stopped")

// 	if s.logsBacklog != nil {
// 		_ = s.logsBacklog.Close()
// 	}
// 	if s.programBacklog != nil {
// 		_ = s.programBacklog.Close()
// 	}

// 	// ✅ esperar que se vacíen todas las goroutines
// 	s.wg.Wait()

// 	close(s.logSocket.Messages)
// 	close(s.programSocket.Messages)
// 	s.QueueManager.Close()

// 	prog := atomic.LoadUint64(&s.programEventsSeen)
// 	logs := atomic.LoadUint64(&s.logEventsSeen)
// 	logrus.Infof("📊 Totals: %d program events, %d log events seen", prog, logs)
// }

func (s *Service) Stop() {
	s.logSocket.Close()
	s.programSocket.Close()
	logrus.Info("Service dependencies stopped")

	if s.logsBacklog != nil {
		_ = s.logsBacklog.Close()
	}
	if s.programBacklog != nil {
		_ = s.programBacklog.Close()
	}

	s.wg.Wait()

	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = s.EventStore.AwaitAsyncPublishes(ctx)
		cancel()
		s.EventStore.Close()
	}

	close(s.logSocket.Messages)
	close(s.programSocket.Messages)
	s.QueueManager.Close()

	prog := atomic.LoadUint64(&s.programEventsSeen)
	logs := atomic.LoadUint64(&s.logEventsSeen)
	logrus.Infof("📊 Totals: %d program events, %d log events seen", prog, logs)
}

// func (s *Service) Stop() {
// 	s.logSocket.Close()
// 	s.programSocket.Close()
// 	logrus.Info("Service dependencies stopped")

// 	if s.logsBacklog != nil {
// 		_ = s.logsBacklog.Close()
// 	}
// 	if s.programBacklog != nil {
// 		_ = s.programBacklog.Close()
// 	}

// 	s.QueueManager.Close()
// 	close(s.logSocket.Messages)
// 	close(s.programSocket.Messages)

// 	prog := atomic.LoadUint64(&s.programEventsSeen)
// 	logs := atomic.LoadUint64(&s.logEventsSeen)
// 	logrus.Infof("📊 Totals: %d program events, %d log events seen", prog, logs)
// }
