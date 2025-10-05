package consumer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"moonmap.io/s3-service/core"
)

type Consumer struct {
	service *core.Service
}

func New() *Consumer {
	return &Consumer{service: core.New()}
}

func (c *Consumer) createTransformConsumer() {
	stream := "media"
	consumer := "media-transform"
	subjects := []string{"media.uploaded.>"}

	c.service.EventStore.CreateConsumer(stream, consumer, subjects,
		func(msg jetstream.Msg) error {
			b := msg.Data()

			mainLog := logrus.WithFields(logrus.Fields{
				"stream":   stream,
				"consumer": consumer,
				"subjects": strings.Join(subjects, ","),
			})

			var media core.MediaState
			if err := json.Unmarshal(b, &media); err != nil {
				mainLog.WithError(err).Error("error while unmarshal core.MediaState")
				return err
			}

			select {
			case c.service.MediaChannel <- media:
			default:
				err := fmt.Errorf("queue full")
				mainLog.Error(err)
				return err
			}
			return nil
		})
}
