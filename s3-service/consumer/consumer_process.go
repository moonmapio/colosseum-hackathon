package consumer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/h2non/bimg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/s3-service/core"
)

func CreateAssetsStatus(status string) map[string]any {
	return map[string]any{
		"$set":         bson.M{"status": status},
		"$currentDate": bson.M{"updatedAt": true},
	}
}

func (c *Consumer) notify(doc *persistence.MediaDoc, status string, ok bool) {
	stream := "notify"
	subject := doc.CreateNotifySubject()
	msgID := doc.CreateMessageId()
	data := core.MediaStateFromDocument(doc)
	data.Status = status
	data.TransitionOk = ok
	err := c.service.EventStore.PublishJSON(stream, subject, msgID, data, nil)
	if err != nil {
		logrus.Error("failed: publishing. Verify connection to NATS server")
	}
}

func (c *Consumer) UpdateAsset(ctx context.Context, doc *persistence.MediaDoc, nextStatus string) {
	mainLog := logrus.WithFields(logrus.Fields{"key": doc.Key})
	mainFilter := map[string]any{"key": doc.Key}
	res, err := c.service.Coll.UpdateOne(ctx, mainFilter, CreateAssetsStatus(nextStatus))
	if err != nil {
		mainLog.WithError(err).Errorf("transition %v->%v failed", doc.Status, nextStatus)
		c.notify(doc, nextStatus, false)
		return
	}

	if res.ModifiedCount == 0 {
		// otro pod ya lo tomó o no está en uploaded
		mainLog.Warn("skipping: not in 'uploaded'")
		return
	}

	doc.Status = nextStatus
	mainLog.Infof("media asset update from  status %v->%v", doc.Status, nextStatus)
	c.notify(doc, nextStatus, true)
}

func (c *Consumer) RemoveObject(ctx context.Context, doc *persistence.MediaDoc) {
	mainLog := logrus.WithFields(logrus.Fields{"key": doc.Key, "s3Bucket": c.service.S3Bucket})
	_, err := c.service.S3c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.service.S3Bucket),
		Key:    aws.String(doc.Key),
	})

	if err != nil {
		mainLog.WithError(err).Error("error while deleting from s3 S3Bucket")
		c.notify(doc, "deleted", false)
		return
	}

	mainLog.Info("media asset removed")
}

func (c *Consumer) process(media core.MediaState) {
	ctx, cancel := context.WithTimeout(c.service.Ctx, 2*time.Minute)
	defer cancel()

	var doc persistence.MediaDoc
	if err := c.service.Coll.FindOne(ctx, bson.M{"key": media.Key}).Decode(&doc); err != nil {
		logrus.WithError(err).WithField("key", media.Key).Error("doc not found")
		return
	}

	c.UpdateAsset(ctx, &doc, "processing")
	// TODO: notify proccesing

	// 2) descarga original
	targetObj := &s3.GetObjectInput{
		Bucket: aws.String(c.service.S3Bucket),
		Key:    aws.String(media.Key),
	}
	obj, err := c.service.S3c.GetObject(ctx, targetObj)
	if err != nil {
		c.UpdateAsset(ctx, &doc, "failed")
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":    doc.Key,
			"bucket": c.service.S3Bucket,
		}).Errorln("error while downloading object from s3 bucket")
		return
	}

	defer obj.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj.Body); err != nil {
		c.UpdateAsset(ctx, &doc, "failed")
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":    doc.Key,
			"bucket": c.service.S3Bucket,
		}).Errorln("error while reading buffer from s3 bucket")
		return
	}

	orig := buf.Bytes()

	// Bytes en memoria: corta si supera tu límite (defensa de doble capa)
	sentBytes := int64(len(orig))
	if sentBytes > c.service.MaxUploadBytes {
		c.UpdateAsset(ctx, &doc, "failed")
		c.RemoveObject(ctx, &doc)
		logrus.WithFields(logrus.Fields{
			"key":   doc.Key,
			"bytes": sentBytes,
		}).Errorln("media asset too big. Will be removed from s3")
		return
	}

	// Dimensiones (px)
	size, err := bimg.NewImage(orig).Size()
	if err != nil || int64(size.Width)*int64(size.Height) > c.service.MaxPixels || size.Width <= 0 || size.Height <= 0 {
		c.UpdateAsset(ctx, &doc, "failed")
		c.RemoveObject(ctx, &doc)
		logrus.WithFields(logrus.Fields{
			"key":  doc.Key,
			"size": size,
		}).Errorln("media asset wrong dimensions. Will be removed from s3")
		return
	}

	sum := sha256.Sum256(orig)
	chk := hex.EncodeToString(sum[:])

	out := make([]persistence.MediaVariant, 0, len(doc.Planned))
	for _, pv := range doc.Planned {

		if !strings.HasPrefix(pv.Mime, "image/") {
			// TODO: gif/video pipeline futura
			continue
		}

		proccessObj := bimg.Options{
			Type:          bimg.WEBP,
			Quality:       80,
			StripMetadata: true,
			Width:         pv.W,
			Height:        pv.H,
			Enlarge:       false,
		}
		img, e := bimg.NewImage(orig).Process(proccessObj)
		if e != nil {
			logrus.WithError(e).WithField("key", pv.Key).Warn("variant process failed")
			continue
		}

		acl := types.ObjectCannedACLPrivate
		if c.service.S3PublicAcl {
			acl = types.ObjectCannedACLPublicRead
		}

		_, e = c.service.S3c.PutObject(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(c.service.S3Bucket),
			Key:          aws.String(pv.Key),
			Body:         bytes.NewReader(img),
			ACL:          acl,
			ContentType:  aws.String(pv.Mime),
			CacheControl: aws.String("public, max-age=31536000, immutable"),
		})

		if e != nil {
			logrus.WithError(e).WithField("key", pv.Key).Warn("put variant failed")
			continue
		}

		vsz, _ := bimg.NewImage(img).Size()
		out = append(out, persistence.MediaVariant{
			Key: pv.Key, W: vsz.Width, H: vsz.Height, Bytes: int64(len(img)),
		})
	}

	orig = nil
	status := "ready"
	if len(out) == 0 {
		status = "failed"
	}

	set := bson.M{
		"checksum":        chk,
		"variants":        out,
		"status":          status,
		"width":           size.Width,
		"height":          size.Height,
		"pipelineVersion": core.PipelineVersion,
	}

	filter := bson.M{"key": doc.Key}
	update := bson.M{
		"$set":         set,
		"$currentDate": bson.M{"updatedAt": true},
	}

	var updated persistence.MediaDoc
	err = c.service.Coll.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logrus.WithError(err).Warn("document not found")
		} else {
			logrus.WithError(err).Error("findOneAndUpdate failed")
		}

		return
	}

	logrus.WithFields(logrus.Fields{
		"key":        doc.Key,
		"uploaderId": media.UploaderID,
	}).Info("media processed successfully")

	doc.Status = status
	c.notify(&updated, status, true)

}
