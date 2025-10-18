package publisher

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"moonmap.io/go-commons/persistence"
)

type TransitionRequest struct {
	Filter bson.M
	Update bson.M
}

func CreateTransitionRequest(
	doc *persistence.MediaDoc,
	nextStatus string,
	expireAt *time.Time,
	bytes *int64,
	mime, etag *string) *TransitionRequest {

	set := bson.M{"status": nextStatus}

	if bytes != nil {
		set["bytes"] = *bytes
		doc.Bytes = *bytes
	}

	if mime != nil {
		doc.Mime = *mime
		set["mime"] = *mime
	}

	if etag != nil {
		set["etag"] = *etag
		doc.ETag = *etag
	}

	if expireAt != nil {
		set["expireAt"] = *expireAt
	}

	update := bson.M{
		"$set":         set,
		"$unset":       bson.M{"nextCheckAt": "", "attempts": ""},
		"$currentDate": bson.M{"updatedAt": true}, // ← timestamp servidor
	}

	filter := bson.M{"key": doc.Key, "status": "pending"} // transición atómica
	return &TransitionRequest{Filter: filter, Update: update}
}
