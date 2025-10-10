package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PresignItemReq struct {
	Mime string `json:"mime"`
	Ext  string `json:"ext,omitempty"`
}

type PresignBatchReq struct {
	Items  []PresignItemReq `json:"items"`
	UserID string           `json:"userId"`
}

type PresignItemRes struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	UploadURL string    `json:"uploadUrl"`
	PublicURL string    `json:"publicUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type PresignBatchRes struct {
	Items []PresignItemRes `json:"items"`
}

type UploadPendingDoc struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	SphereID  string        `bson:"sphereId"`
	UserID    bson.ObjectID `bson:"userId"`
	Key       string        `bson:"key"`
	Mime      string        `bson:"mime"`
	Status    string        `bson:"status"`
	ExpiresAt time.Time     `bson:"expiresAt"`
	CreatedAt time.Time     `bson:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}
