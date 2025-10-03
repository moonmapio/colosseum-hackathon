package system

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

type S3Config struct {
	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string
	S3Bucket    string
	S3Region    string
	S3PublicAcl bool
}

func LoadS3(ctx context.Context) (*S3Config, *s3.Client, *s3.PresignClient) {
	s3Config := &S3Config{
		S3AccessKey: helpers.GetEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: helpers.GetEnv("S3_SECRET_KEY", ""),
		S3Endpoint:  helpers.GetEnv("S3_ENDPOINT", "https://fsn1.your-objectstorage.com"),
		S3Bucket:    helpers.GetEnv("S3_BUCKET", "moonmap"),
		S3Region:    helpers.GetEnv("S3_REGION", "eu-central"),
		S3PublicAcl: helpers.GetEnv("S3_PUBLIC_ACL", "true") == "true",
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s3Config.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Config.S3AccessKey, s3Config.S3SecretKey, "")),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = 3
			})
		}),
	)

	if err != nil {
		logrus.Fatal(err)
	}

	S3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = false
		o.BaseEndpoint = aws.String(s3Config.S3Endpoint) // endpoint regional (no el del bucket)
	})

	Presigner := s3.NewPresignClient(S3c)

	return s3Config, S3c, Presigner
}
