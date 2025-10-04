package service

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
)

func (s *Service) CreateConsumerSpheres() {
	hostname, _ := os.Hostname()
	hostname = strings.Split(hostname, ".")[0]

	consumerName := hostname

	stream := "spheres"
	subjects := []string{"spheres.>"} // todos los eventos bajo spheres

	s.EventStore.CreateConsumer(stream, consumerName, subjects,
		func(msg jetstream.Msg) error {
			data := msg.Data()
			logrus.Infof("consuming event subject=%s", msg.Subject())

			// forward al Hub
			var ev map[string]any
			if err := json.Unmarshal(data, &ev); err != nil {
				return err
			}

			s.Hub.BroadcastJSON(msg.Subject(), ev)
			return nil
		})
}
