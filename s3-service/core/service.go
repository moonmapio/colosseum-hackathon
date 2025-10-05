package core

import (
	"context"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/h2non/bimg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
)

type Service struct {
	Ctx         context.Context
	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string
	S3Bucket    string
	S3Region    string
	S3PublicAcl bool

	Coll      *mongo.Collection
	S3c       *s3.Client
	Presigner *s3.PresignClient

	MediaChannel chan MediaState
	Workers      int

	MaxUploadBytes int64
	MaxPixels      int64
	AllowedMimes   map[string]bool

	EventStore *system.NatsEventStore
	Mode       string
}

func New() *Service {

	mode := helpers.GetEnv("MODE", "")
	if mode == "" {
		logrus.Panicln("MODE not set")
	}

	// system.LoadS3()

	w, _ := strconv.Atoi(helpers.GetEnv("WORKERS", "2"))
	s := &Service{
		S3AccessKey:  helpers.GetEnvOrFail("S3_ACCESS_KEY"),
		S3SecretKey:  helpers.GetEnvOrFail("S3_SECRET_KEY"),
		S3Endpoint:   helpers.GetEnv("S3_ENDPOINT", "https://fsn1.your-objectstorage.com"),
		S3Bucket:     helpers.GetEnv("S3_BUCKET", "moonmap"),
		S3Region:     helpers.GetEnv("S3_REGION", "eu-central"),
		S3PublicAcl:  helpers.GetEnv("S3_PUBLIC_ACL", "true") == "true",
		Workers:      w,
		MediaChannel: make(chan MediaState, 1024),
	}

	s.Mode = mode

	maxBytes, _ := strconv.ParseInt(helpers.GetEnv("MAX_UPLOAD_BYTES", "10485760"), 10, 64) // 10MB
	maxPx, _ := strconv.ParseInt(helpers.GetEnv("MAX_PIXELS", "25000000"), 10, 64)          // 25MP
	mimes := strings.Split(helpers.GetEnv("ALLOWED_MIME", "image/png,image/jpeg,image/webp"), ",")
	allow := make(map[string]bool, len(mimes))
	for _, m := range mimes {
		allow[strings.ToLower(strings.TrimSpace(m))] = true
	}

	s.MaxUploadBytes = maxBytes
	s.MaxPixels = maxPx
	s.AllowedMimes = allow

	return s
}

func (s *Service) Config(ctx context.Context) {
	s.Ctx = ctx

	indexes := CreateMongoIndexes()
	s.Coll = persistence.MustGetCollection(constants.MediaAssetsCollectionName)
	_, err := s.Coll.Indexes().CreateMany(s.Ctx, indexes)
	if err != nil {
		logrus.Fatal(err)
	}

	if s.IsConsumer() {
		bimg.VipsCacheSetMaxMem(32 * 1024 * 1024)
		bimg.VipsCacheSetMax(100)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.S3AccessKey, s.S3SecretKey, "")),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = 3
			})
		}),
	)

	if err != nil {
		logrus.Fatal(err)
	}

	s.S3c = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = false
		o.BaseEndpoint = aws.String(s.S3Endpoint) // endpoint regional (no el del bucket)
	})

	s.Presigner = s3.NewPresignClient(s.S3c)

	s.EventStore = system.NewEventStore("s3-service")

	logrus.Infof("s3 client configured successfully in mode %v", s.Mode)
}

func (s *Service) IsConsumer() bool {
	return s.Mode == "consumer"
}

func IsConsumer() bool {
	mode := helpers.GetEnv("MODE", "")
	return mode == "consumer"
}

func (s *Service) IsPublisher() bool {
	return s.Mode == "publisher"
}

func IsPublisher() bool {
	mode := helpers.GetEnv("MODE", "")
	return mode == "publisher"
}
