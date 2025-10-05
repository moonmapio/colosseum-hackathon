package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
)

type NatsEventStore struct {
	conn                 *nats.Conn
	js                   *jetstream.JetStream
	streams              map[string]*jetstream.Stream
	consumers            map[string]*jetstream.Consumer
	subscriptions        map[string]*nats.Subscription
	DisconnectErrHandler func(nc *nats.Conn, err error)
	ReconnectHandler     func(nc *nats.Conn)
	ClosedHandler        func(nc *nats.Conn)
}

func NewEventStore(name string) *NatsEventStore {

	store := &NatsEventStore{
		subscriptions: map[string]*nats.Subscription{},
		streams:       make(map[string]*jetstream.Stream),
		consumers:     make(map[string]*jetstream.Consumer),
	}

	logrus.Info("📦 Creating NATS EventStore...s")

	url := helpers.GetEnv("NATS_URL", "127.0.0.1:4222")
	if !strings.HasPrefix(url, "nats://") && !strings.HasPrefix(url, "tls://") {
		url = "nats://" + url
	}

	conn, err := nats.Connect(
		url,
		nats.Name(name),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.Timeout(5*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logrus.Warnf("[%s] disconnected from NATS: %v", name, err)
			if store.DisconnectErrHandler != nil {
				store.DisconnectErrHandler(nc, err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logrus.Infof("[%s] reconnected to %s", name, nc.ConnectedUrl())
			if store.ReconnectHandler != nil {
				store.ReconnectHandler(nc)
			}
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logrus.Warnf("[%s] NATS connection closed", name)
			if store.ClosedHandler != nil {
				store.ClosedHandler(nc)
			}
		}),
	)
	if err != nil {
		logrus.Fatalf("❌ Failed to connect to NATS at %s: %v", url, err)
	}

	js, err := jetstream.New(conn, jetstream.WithPublishAsyncMaxPending(10000))
	if err != nil {
		logrus.Fatalf("❌ Failed to connect to JetStream at %s: %v", url, err)
	}

	logrus.Infof("✅ Connected to NATS at %s\n", url)
	store.conn = conn
	store.js = &js
	return store
}

func (n *NatsEventStore) GetConn() *nats.Conn {
	return n.conn
}

func (n *NatsEventStore) CreateStreamWithSubjects(ctx context.Context, streamName string, subjects []string) {
	if len(subjects) == 0 {
		subjects = []string{fmt.Sprintf("%v.>", streamName)}
	}

	config := jetstream.StreamConfig{
		Name:        streamName,
		Retention:   jetstream.LimitsPolicy,
		Subjects:    subjects,
		Storage:     jetstream.FileStorage,
		Duplicates:  2 * time.Hour,
		AllowRollup: true,
		Replicas:    1,
		MaxAge:      7 * 24 * time.Hour,
		MaxBytes:    1073741824, // 1GB
	}

	n.CreateStreamWithConfig(ctx, config)
}

func (n *NatsEventStore) CreateStreamWithConfig(ctx context.Context, config jetstream.StreamConfig) {
	localCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	stream, err := (*n.js).CreateOrUpdateStream(localCtx, config)
	if err != nil {
		logrus.WithField("streamName", config.Name).Error(err)
		return
	}

	n.streams[config.Name] = &stream
	logrus.Infof("stream %v created or updated successfully", config.Name)
}

func (n *NatsEventStore) CreateEphemeralConsumer(streamName, consumerName string, subjects []string, handler func(msg jetstream.Msg) error) {
	n.createConsumerInternal(streamName, consumerName, subjects, handler, false)

}
func (n *NatsEventStore) CreateConsumer(streamName, consumerName string, subjects []string, handler func(msg jetstream.Msg) error) {
	n.createConsumerInternal(streamName, consumerName, subjects, handler, true)
}

func (n *NatsEventStore) createConsumerInternal(streamName, consumerName string, subjects []string, handler func(msg jetstream.Msg) error, durable bool) {
	ctx := context.Background()
	config := jetstream.ConsumerConfig{
		Name:           consumerName,
		FilterSubjects: subjects,
		AckWait:        1 * time.Minute,
		AckPolicy:      jetstream.AckExplicitPolicy,
		DeliverPolicy:  jetstream.DeliverNewPolicy,
		ReplayPolicy:   jetstream.ReplayInstantPolicy,
		MaxDeliver:     12,
		MaxAckPending:  1024,
		BackOff: []time.Duration{
			5 * time.Second,
			15 * time.Second,
			30 * time.Second,
			2 * time.Minute,
			5 * time.Minute,
		},
	}

	if durable {
		config.Durable = consumerName
	}

	consumer, err := (*n.js).CreateOrUpdateConsumer(ctx, streamName, config)
	if err != nil {
		logrus.WithField("streamName", streamName).Panic(err)
		return
	}

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		attempt := 1
		meta, _ := msg.Metadata()
		subject := msg.Subject()
		fmt.Println(subject)
		if meta != nil {
			attempt = int(meta.NumDelivered) // 1,2,3,...
		}
		err := handler(msg)
		if err == nil {
			_ = msg.Ack()
			return
		}

		// Mapea tu sentinel de “no listo todavía”
		if errors.Is(err, constants.ErrNotReady) {
			// elige delay desde cfg.BackOff según attempt
			delay := config.AckWait
			if len(config.BackOff) > 0 {
				idx := attempt - 1
				if idx >= len(config.BackOff) {
					idx = len(config.BackOff) - 1
				}
				delay = config.BackOff[idx]
			}

			logrus.WithFields(logrus.Fields{
				"consumer": consumerName,
				"subject":  strings.Join(subjects, ","),
				"attempt":  attempt,
				"delay":    delay.String(),
			}).Info("NAK with delay (backoff)")

			_ = msg.NakWithDelay(delay)
			return
		}

		// Para errores “reales”, puedes NAK inmediato o TERM
		logrus.WithError(err).WithFields(logrus.Fields{
			"consumer": consumerName,
			"attempt":  attempt,
		}).Warn("NAK (immediate) due to error")
		_ = msg.Nak()
	})

	if err != nil {
		logrus.WithField("consumerName", consumerName).WithError(err).Error("consume failed")
		return
	}

	n.consumers[consumerName] = &consumer
	fields := logrus.Fields{"streamName": streamName, "consumerName": consumerName, "subjects": strings.Join(subjects, ", ")}
	logrus.WithFields(fields).Info("consumer created")
}

func (n *NatsEventStore) Close() {
	for _, sub := range n.subscriptions {
		if sub != nil {
			_ = sub.Unsubscribe()
		}
	}
	if n.conn != nil {
		_ = n.conn.Drain()
		n.conn.Close()
	}
}

func (n *NatsEventStore) PublishStream(ctx context.Context, stream, eventType string, eventData any) error {
	data, err := helpers.Encode(eventData)
	if err != nil {
		return err
	}

	ackFuture, err := (*n.js).PublishAsync(eventType, data, jetstream.WithExpectStream(stream))
	if err != nil {
		logrus.WithField("streamName", stream).WithError(err).Error("Failed to publish message async")
		return err
	}

	go n.handleAsyncAck(eventType, ackFuture)
	return nil
}

func (n *NatsEventStore) PublishJSON(stream, subject, msgID string, data any, hdr nats.Header) error {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = n.PublishBytes(stream, subject, msgID, bytesData, hdr)
	return err
}

func (n *NatsEventStore) PublishBytes(stream, subject, msgID string, bytesData []byte, hdr nats.Header) (jetstream.PubAckFuture, error) {
	msg := nats.NewMsg(subject)

	if hdr == nil {
		msg.Header = nats.Header{}
	} else {
		msg.Header = nats.Header{}
		for k, vals := range hdr {
			for _, vv := range vals {
				msg.Header.Add(k, vv)
			}
		}
	}

	if msgID != "" {
		msg.Header.Set("Nats-Msg-Id", msgID)
	}

	msg.Data = bytesData
	pa, err := (*n.js).PublishMsgAsync(msg, jetstream.WithExpectStream(stream))
	if err != nil {
		logrus.WithField("streamName", stream).WithError(err).Error("Failed to publish message async")
		return nil, err
	}

	// go n.handleAsyncAck(subject, pa)
	return pa, nil
}

func (n *NatsEventStore) PublishBytesSync(stream, subject, msgID string, bytesData []byte, hdr nats.Header) error {
	msg := nats.NewMsg(subject)

	if hdr != nil {
		msg.Header = hdr
	} else {
		msg.Header = nats.Header{}
	}
	if msgID != "" {
		msg.Header.Set("Nats-Msg-Id", msgID)
	}
	msg.Data = bytesData

	// Publicación sincrónica → espera el PubAck
	_, err := (*n.js).PublishMsg(context.Background(), msg, jetstream.WithExpectStream(stream))
	if err != nil {
		logrus.WithField("streamName", stream).WithError(err).Error("Failed to publish message sync")
		return err
	}
	return nil
}

func (n *NatsEventStore) AwaitAsyncPublishes(ctx context.Context) error {
	select {
	case <-(*n.js).PublishAsyncComplete():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for async publishes")
	}
}

func (n *NatsEventStore) handleAsyncAck(subject string, ackFuture jetstream.PubAckFuture) {
	select {
	case ack := <-ackFuture.Ok():
		logrus.WithFields(logrus.Fields{
			"eventType": subject,
			"stream":    ack.Stream,
			"sequence":  ack.Sequence,
		}).Debug("Message published to JetStream (async)")
	case err := <-ackFuture.Err():
		logrus.WithField("eventType", subject).WithError(err).Info("Async publish failed")
	case <-time.After(5 * time.Second):
		logrus.WithField("eventType", subject).Info("Async publish timeout")
	}
}
