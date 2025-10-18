package service

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
)

type Service struct {
	Ctx        context.Context
	EventStore *system.NatsEventStore
	Coll       *mongo.Collection
}

func New() *Service {
	service := Service{}
	service.EventStore = system.NewEventStore(constants.RequestRecorderServiceName)
	return &service
}

func (s *Service) Config(ctx context.Context) {
	s.Ctx = ctx
	s.Coll = persistence.MustGetCollection(constants.RequestCollectionName)
	s.createConsumer()
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx)
		<-ctx.Done()
	})
}
