package system

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

var instance *System
var once sync.Once

type System struct {
	startTime      time.Time
	elapsedTime    time.Duration
	ctx            context.Context
	cancelFunc     context.CancelFunc
	mu             sync.Mutex
	cleanUpHooks   []func()
	shutdownClosed bool
	Bind           string
}

func New() *System {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		bind := helpers.GetEnv("BIND", ":8080")
		instance = &System{ctx: ctx, cancelFunc: cancel, startTime: time.Now(), Bind: bind}
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			instance.Shutdown()
		}()
	})
	return instance
}

func GetInstance() *System          { return New() }
func GetContext() context.Context   { return GetInstance().ctx }
func GetCancel() context.CancelFunc { return GetInstance().cancelFunc }

func (s *System) GetCancel() context.CancelFunc {
	return s.cancelFunc
}

func (s *System) LogElapsedTime() {
	s.elapsedTime = time.Since(s.startTime)
	logrus.WithField("elapsed", s.elapsedTime.Minutes()).Info("⏱ process finalized")
}

func (s *System) SetFormatter() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.AddHook(&PidHook{Pid: os.Getpid()})

	if os.Getenv("LOG_FILE") == "true" {
		filename := "app.log"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			f, err := os.Create(filename)
			if err != nil {
				logrus.Fatalf("no se pudo crear archivo de log: %v", err)
			}
			f.Close()
		}
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logrus.Fatalf("no se pudo abrir archivo de log: %v", err)
		}
		logrus.SetOutput(f)

		// opcional: timestamp legible
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

}

func (s *System) Shutdown() {
	if !s.shutdownClosed {
		s.shutdownClosed = true
		s.cancelFunc()
		s.runCleanUpHooks()
		s.LogElapsedTime()
	}
}

func (s *System) Run(runnable func(ctx context.Context)) {
	runnable(s.ctx)
}
