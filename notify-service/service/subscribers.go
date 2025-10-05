package service

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

// Suscripción global a notify.> (sin queue group)
func (s *Service) CreateSubscriberNotify() {
	_, err := s.EventStore.GetConn().Subscribe("notify.>", func(m *nats.Msg) {
		msgID := m.Header.Get("Nats-Msg-Id")
		var streamSeq string
		if md, e := m.Metadata(); e == nil {
			streamSeq = fmt.Sprintf("%d", md.Sequence.Stream)
		} else {
			streamSeq = ksuid.New().String()
		}

		logrus.Infof("new message recived from on subject=%v with id=%v seq=%s", m.Subject, msgID, streamSeq)

		s.Hub.BroadcastJSON(m.Subject, NotifyEvent{
			Subject: m.Subject,
			ID:      fmt.Sprintf("%v:%v", msgID, streamSeq),
			Data:    m.Data,
			Header:  m.Header,
		})
	})

	if err != nil {
		logrus.Panic(err)
		return
	}

	logrus.Info("subscribed to notify.>")
}
