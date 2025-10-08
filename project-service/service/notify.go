package service

import (
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/persistence"
)

func (s *Service) notify(doc *persistence.ProjectDoc, status string) {
	subject := doc.CreateNotifySubject(status)
	msgID := doc.CreateMessageId()
	err := s.EventStore.PublishJSON(constants.StreamNotify, subject, msgID, doc, nil)
	if err != nil {
		logrus.Error("failed: publishing. Verify connection to NATS server")
	}
}
