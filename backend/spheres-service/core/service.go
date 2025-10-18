package core

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
)

type Service struct {
	ctx                context.Context
	spheresColl        *mongo.Collection
	mediaColl          *mongo.Collection
	sphereContentsColl *mongo.Collection

	EventStore *system.NatsEventStore

	S3Cfg     *system.S3Config
	S3c       *s3.Client
	Presigner *s3.PresignClient
}

func NewService() *Service {
	s := &Service{}
	return s
}

func (s *Service) Config() {
	s.spheresColl = persistence.MustGetCollection(constants.SpheresCollectionName)
	s.sphereContentsColl = persistence.MustGetCollection(constants.SphereContentsCollectionName)
	s.mediaColl = persistence.MustGetCollection(constants.SphereMediaCollectionName)

	// Indexes básicos
	_, err := s.spheresColl.Indexes().CreateMany(s.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "mint", Value: 1}}, Options: options.Index().SetUnique(true)},
	})

	if err != nil {
		logrus.Fatal(err)
	}

	_, err = s.sphereContentsColl.Indexes().CreateMany(s.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "sphereId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "parentId", Value: 1}}},
	})

	if err != nil {
		logrus.Fatal(err)
	}

	_, err = s.mediaColl.Indexes().CreateMany(s.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "sphereId", Value: 1}, {Key: "userId", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
	})

	if err != nil {
		logrus.Fatal(err)
	}

	s.EventStore = system.NewEventStore(constants.SpheresServiceName)

	s.S3Cfg, s.S3c, s.Presigner = system.LoadS3(s.ctx)
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.ctx = ctx
		s.Config()

		// go s.createConsumer()
		ownhttp.NewServer(ctx, constants.SpheresServiceName, sys.Bind, s.routes(), nil)
		<-ctx.Done()
	})
}
