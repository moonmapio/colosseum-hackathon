package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
)

func (s *Service) CreateConsumerSpheres() {
	hostname, _ := os.Hostname()
	hostname = strings.Split(hostname, ".")[0]
	consumerName := hostname

	subjects := []string{fmt.Sprintf("%s.>", constants.StreamSpheres)}
	s.EventStore.CreateConsumer(constants.StreamSpheres, consumerName, subjects,
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
