package service

import (
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"moonmap.io/go-commons/constants"
)

func (s *Service) createConsumer() {
	consumer := "request-recorder-consumer"
	subjects := []string{"requests.incomming"}

	s.EventStore.CreateConsumer(constants.StreamRequests, consumer, subjects,
		func(msg jetstream.Msg) error {
			data := msg.Data()
			mainLog := logrus.WithFields(logrus.Fields{
				"stream":   constants.StreamRequests,
				"consumer": consumer,
				"subjects": strings.Join(subjects, ","),
			})
			mainLog.Info("consuming event coming from own service type")

			mainLog.WithField("raw", string(data)).Info("received raw event")

			var doc bson.Raw
			if err := bson.UnmarshalExtJSON(data, true, &doc); err != nil {
				mainLog.WithError(err).Error("error converting JSON to BSON")
				return err // NAK
			}

			_, err := s.Coll.InsertOne(s.Ctx, doc)
			if err != nil {
				mainLog.WithError(err).Error("error inserting request into MongoDB")
				return err // NAK
			}

			mainLog.Info("request saved into MongoDB as raw BSON")
			return nil // ACK
		},
	)
}
