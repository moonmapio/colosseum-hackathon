package service

import (
	"context"
	"time"

	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/messages"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
)

type Service struct {
	startedAt    time.Time
	Ctx          context.Context
	CancelFunc   context.CancelFunc
	QueueManager *messages.QueueManager
}

func New() *Service {
	s := &Service{}

	if s.startedAt.IsZero() {
		s.startedAt = time.Now().UTC()
	}

	s.QueueManager = messages.NewManager(constants.AlertServiceName, "support@moonmap.io", "MoonMap Support", s.startedAt)
	return s
}

func (s *Service) Config(ctx context.Context, cancelFunc context.CancelFunc) {
	s.Ctx = ctx
	s.CancelFunc = cancelFunc

	s.QueueManager.SetContext(s.Ctx)
	s.QueueManager.StartIdleAggregator(10 * time.Second)

}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx, sys.GetCancel())
		ownhttp.NewServer(ctx, constants.NotifyServiceName, sys.Bind, s.routes(), nil)
		<-ctx.Done()
		s.QueueManager.Close()
	})
}
