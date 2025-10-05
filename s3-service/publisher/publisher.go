package publisher

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/s3-service/core"
)

type Publisher struct {
	service *core.Service
}

func New() *Publisher {
	return &Publisher{service: core.New()}
}

func (p *Publisher) createPendingConsumer() {
	stream := "media"
	consumer := "media-publisher"
	subjects := []string{"media.pending"}

	p.service.EventStore.CreateConsumer(stream, consumer, subjects,
		func(msg jetstream.Msg) error {
			data := msg.Data()
			mainLog := logrus.WithFields(logrus.Fields{
				"stream":   stream,
				"consumer": consumer,
				"subjects": strings.Join(subjects, ","),
			})
			mainLog.Info("consuming event coming from own service type")

			var ev core.MediaState
			if err := json.Unmarshal(data, &ev); err != nil {
				mainLog.WithError(err).Error("error while unmarshal core.MediaState")
				return err // NAK
			}

			// buscar el documento
			var doc persistence.MediaDoc
			err := p.service.Coll.
				FindOne(p.service.Ctx, bson.M{"key": ev.Key, "status": "pending"}).
				Decode(&doc)

			if err != nil {
				if err == mongo.ErrNoDocuments {
					mainLog.WithError(err).WithFields(logrus.Fields{
						"key": ev.Key,
					}).Warn("pending consumer: doc not found or not pending")
					// Doc aún no insertado o ya cambió de estado → ACK o NAK?
					// Recomendado: NAK suave para reintentar.
					return constants.ErrNotReady
				}

				mainLog.WithError(err).WithFields(logrus.Fields{
					"key": ev.Key,
				}).Error("unable to find any document for request key")
				return err // NAK por error real
			}

			return p.tryConfirm(doc)
		},
	)
}

func (p *Publisher) tryConfirm(doc persistence.MediaDoc) error {
	mainLog := logrus.WithFields(logrus.Fields{
		"key":             doc.Key,
		"entityId":        doc.EntityID,
		"status":          doc.Status,
		"attempts":        doc.Attempts,
		"pipelineVersion": doc.PipelineVersion,
	})

	mainLog.Info("confirm: HEAD check started")

	ctx, cancel := context.WithTimeout(p.service.Ctx, 3*time.Second)
	defer cancel()

	// HEAD
	ok, mime, bytes, etag := p.headOkInternal(ctx, doc.Key)
	if !ok {
		// NAK → backoff y reintento automático
		mainLog.Warn("file not present in bucket yet. Retry follows")
		return constants.ErrNotReady
	}

	// Transición atómica pending→uploaded
	req := CreateTransitionRequest(&doc, "uploaded", nil, &bytes, &mime, &etag)
	res, err := p.service.Coll.UpdateOne(p.service.Ctx, req.Filter, req.Update)
	if err != nil {
		mainLog.WithError(err).Error("confirm: db transition pending→uploaded failed")
		return err // NAK
	}

	if res.ModifiedCount == 0 {
		// Ya no está "pending" (otro pod lo movió) → ACK
		mainLog.Warn("confirm: skipped transition (not pending anymore)")
		return nil
	}

	return p.PublishMedia(&doc)
}

func (p *Publisher) PublishMedia(doc *persistence.MediaDoc) error {
	stream := "media"
	// Idempotencia del publish (cualquier string estable sirve)
	msgId := doc.CreateMessageId()
	subject := doc.CreateMediaSubject("uploaded")
	eventData := core.MediaStateFromDocument(doc)
	eventData.Status = "uploaded"

	mainLog := logrus.WithFields(logrus.Fields{
		"key":     doc.Key,
		"subject": subject,
		"stream":  stream,
		"msgId":   msgId,
	})

	err := p.service.EventStore.PublishJSON(stream, subject, msgId, eventData, nil)
	if err != nil {
		mainLog.Error("failed: publishing. Verify connection to NATS server")
		// Publicación falló → NAK para reintentar
		return err

	}

	p.PublishNotify(doc)
	mainLog.Info("confirm: uploaded (pending→uploaded)")
	// ACK
	return nil
}

func (p *Publisher) PublishNotify(doc *persistence.MediaDoc) {
	// 3) (opcional pero muy útil) evento de NOTIFY con rollup por mediaId
	//    Esto te permite a la UI suscribirse a "notify.media.*" y ver siempre el último estado.
	hdr := nats.Header{}
	hdr.Set("Nats-Rollup", "sub") // JetStream: mantén solo el último mensaje por subject
	now := time.Now()

	subject := doc.CreateNotifySubject() // p.ej. notify.media.users.u1.avatar.u1.v1.original_png
	eventData := core.MediaStateFromDocument(doc)
	eventData.Status = "uploaded"
	eventData.UpdatedAt = now

	stream := "notify"
	msgId := doc.CreateMessageId()
	err := p.service.EventStore.PublishJSON(stream, subject, msgId, eventData, hdr)

	mainLog := logrus.WithFields(logrus.Fields{
		"key":     doc.Key,
		"subject": subject,
		"stream":  stream,
		"msgId":   msgId,
	})

	if err != nil {
		mainLog.Error("failed: publishing. Verify connection to NATS server")
		return
	}

	mainLog.Info("confirm: published")
}

func (p *Publisher) headOkInternal(ctx context.Context, key string) (bool, string, int64, string) {
	head, err := p.service.S3c.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.service.S3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logrus.WithError(err).WithField("key", key).Warn("HEAD failed")
		return false, "", 0, ""
	}

	length := aws.ToInt64(head.ContentLength)
	if length <= 0 || length > p.service.MaxUploadBytes {
		logrus.WithFields(logrus.Fields{"key": key, "len": length}).Warn("HEAD reject: invalid length")
		return false, "", 0, ""
	}

	ct := helpers.NormCT(aws.ToString(head.ContentType))
	if !p.service.AllowedMimes[ct] {
		logrus.WithFields(logrus.Fields{"key": key, "ct": ct}).Warn("HEAD reject: unsupported mime")
		return false, "", 0, ""
	}

	etag := strings.Trim(aws.ToString(head.ETag), "\"")
	return true, ct, length, etag
}
