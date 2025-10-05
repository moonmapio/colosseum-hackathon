package service

import (
	"context"
	"time"

	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/messages"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
)

type Service struct {
	startedAt    time.Time
	Ctx          context.Context
	CancelFunc   context.CancelFunc
	EventStore   *system.NatsEventStore
	Origin       string
	Hub          *ownhttp.Hub
	QueueManager *messages.QueueManager
}

func New() *Service {
	// TODO: add origin or list of origins to allow
	allowOrigins := helpers.GetEnv("ALLOW_ORIGIN", "*")
	s := &Service{
		Origin: allowOrigins,
		Hub:    ownhttp.NewHub(),
	}

	if s.startedAt.IsZero() {
		s.startedAt = time.Now().UTC()
	}

	serviceName := constants.NotifyServiceName
	s.QueueManager = messages.NewManager(serviceName, "support@moonmap.io", "MoonMap Support", s.startedAt)

	s.Hub.Mode = "subjects"
	s.EventStore = system.NewEventStore(constants.NotifyServiceName)
	return s
}

func (s *Service) Config(ctx context.Context, cancelFunc context.CancelFunc) {
	s.Ctx = ctx
	s.CancelFunc = cancelFunc

	s.QueueManager.SetContext(s.Ctx)
	s.QueueManager.StartAggregator(5 * time.Second)

	s.CreateStreamMedia()
	s.CreateStreamNotify()
	s.CreateStreamRequest()
	s.CreateStreamSolana()

	s.CreateConsumerSpheres()
	s.CreateSubscriberNotify()
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx, sys.GetCancel())

		opts := ownhttp.ServerOpts{
			Read:  ownhttp.Dur(0),
			Write: ownhttp.Dur(0),
			Idle:  ownhttp.Dur(0),
		}

		ownhttp.NewServer(ctx, constants.NotifyServiceName, sys.Bind, s.routes(), &opts)
		<-ctx.Done()
		s.QueueManager.Close()
		s.Hub.Close()
	})
}
