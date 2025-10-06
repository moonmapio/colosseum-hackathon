package service

import (
	"context"
	"net/http"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
)

type Service struct {
	collEvents   *mongo.Collection
	collSessions *mongo.Collection
	collWallets  *mongo.Collection

	store *Store

	ctx context.Context

	livekitApiKey    string
	livekitApiSecret string
	livekitURL       string
	livekitClient    *lksdk.RoomServiceClient
}

func New() *Service { return &Service{} }

func (s *Service) Config(ctx context.Context) {
	s.collEvents = persistence.MustGetCollection("waves_events")
	s.collSessions = persistence.MustGetCollection("waves_sessions")
	s.collWallets = persistence.MustGetCollection("wallets")

	s.store = NewStore(s.collEvents, s.collSessions, s.collWallets)

	s.livekitApiKey = helpers.GetEnvOrFail("LIVEKIT_API_KEY")
	s.livekitApiSecret = helpers.GetEnvOrFail("LIVEKIT_API_SECRET")
	s.livekitURL = helpers.GetEnvOrFail("LIVEKIT_URL")

	s.ctx = ctx
	_, err := s.collEvents.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "evt.room", Value: 1}, {Key: "at", Value: -1}}},
		{Keys: bson.D{{Key: "evt.event", Value: 1}, {Key: "at", Value: -1}}},
	})

	if err != nil {
		logrus.Fatal(err)
	}

	_, err = s.collSessions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "sphereId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "active", Value: 1}, {Key: "lastUpdated", Value: -1}}},
	})

	if err != nil {
		logrus.Fatal(err)
	}

	s.livekitClient = lksdk.NewRoomServiceClient(s.livekitURL, s.livekitApiKey, s.livekitApiSecret)
	logrus.Infof("%s configured", constants.WavesServiceName)
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx)
		ownhttp.NewServer(ctx, constants.WavesServiceName, sys.Bind, s.routes(), nil)
		<-ctx.Done()
	})
}

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()
	mux.HandleFunc("/waves/join", s.handleJoin)
	mux.HandleFunc("/waves/webhook", s.handleWebhook)
	mux.HandleFunc("/waves/stats", s.handleStats)
	return mux
}
