package service

import (
	"context"
	"net/http"
	"strings"

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
	ctx        context.Context
	coll       *mongo.Collection
	mediaColl  *mongo.Collection
	EventStore *system.NatsEventStore

	S3Cfg     *system.S3Config
	S3c       *s3.Client
	Presigner *s3.PresignClient
}

func New() *Service {
	return &Service{}
}

func (s *Service) Config(ctx context.Context) {
	s.ctx = ctx
	s.coll = persistence.MustGetCollection(constants.ProjectsCollectionName)
	s.mediaColl = persistence.MustGetCollection(constants.MediaAssetsCollectionName)

	s.EventStore = system.NewEventStore(constants.ProjectServiceName)

	cfg, s3c, p := system.LoadS3(ctx)
	s.S3c = s3c
	s.S3Cfg = cfg
	s.Presigner = p

	// Indexes básicos
	_, err := s.coll.Indexes().CreateMany(s.ctx, []mongo.IndexModel{
		// {Keys: bson.D{{Key: "symbol", Value: 1}, {Key: "chain", Value: 1}}, Options: options.Index().SetUnique(false)},
		{Keys: bson.D{{Key: "contractAddress", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
		// {Keys: bson.D{{Key: "createdAt", Value: -1}}},
	})

	if err != nil {
		logrus.Fatal(err)
	}
}

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)
		if ownhttp.IsOptionsMethod(r, w) {
			return
		}

		if r.Method == http.MethodDelete {
			s.HandleRemoveProject(w, r)
			return
		}

		s.HandleCreateOrUpdateProject(w, r)
	})

	mux.HandleFunc("/project/", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		ok := len(parts) == 3 && parts[0] == "project" && parts[2] == "removeLogo"

		if !ok {
			ownhttp.WriteJSONError(w, http.StatusNotFound, "NOT_FOUND", "path not found")
			return
		}

		if ownhttp.IsOptionsMethod(r, w) {
			return
		}

		if r.Method != http.MethodDelete {
			ownhttp.WriteJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}

		projectID := parts[1]
		s.HandleRemoveMediaForProject(w, r, projectID)

	})

	return mux
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx)
		// http server
		ownhttp.NewServer(ctx, constants.ProjectServiceName, sys.Bind, s.routes(), nil)
		<-ctx.Done()
	})
}
