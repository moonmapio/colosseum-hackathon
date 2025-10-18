package main

import (
	"math/rand"
	"time"

	"github.com/h2non/bimg"
	"moonmap.io/go-commons/system"
	"moonmap.io/s3-service/consumer"
	"moonmap.io/s3-service/core"
	"moonmap.io/s3-service/publisher"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	bimg.Initialize()
	sys := system.New()

	sys.LoadEnvFile()
	sys.SetFormatter()

	defer sys.Shutdown()
	defer bimg.Shutdown()

	if core.IsConsumer() {
		consumer := consumer.New()
		consumer.Start(sys)
	} else {
		publisher := publisher.New()
		publisher.Start(sys)
	}
}
