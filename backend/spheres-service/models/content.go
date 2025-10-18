package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SphereContentEmbeddedUser struct {
	ID        bson.ObjectID `bson:"_id" json:"_id"`
	FullName  string        `bson:"fullName" json:"fullName"`
	AvatarUrl string        `bson:"avatarUrl" json:"avatarUrl"`
	// todo add verified
}

type SphereContentCreated struct {
	ID        bson.ObjectID             `bson:"_id" json:"_id"`
	ParentID  *bson.ObjectID            `bson:"parentId,omitempty" json:"parentId"`
	SphereID  string                    `bson:"sphereId" json:"sphereId"`
	User      SphereContentEmbeddedUser `bson:"user" json:"user"`
	Reactions map[string][]string       `bson:"reactions" json:"reactions"`
	Type      string                    `bson:"type" json:"type"`
	Text      string                    `bson:"text" json:"text"`
	MediaUrls []string                  `bson:"mediaUrls" json:"mediaUrls"`
	CreatedAt time.Time                 `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time                 `bson:"updatedAt" json:"updatedAt"`
	Deleted   bool                      `bson:"deleted" json:"deleted"`
}

type SphereContentUpdated struct {
	ID        string                 `json:"_id"`
	SphereID  string                 `json:"sphereId"`
	Updates   map[string]interface{} `json:"updates"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

type SphereContentDeleted struct {
	ID        string    `json:"_id"`
	SphereID  string    `json:"sphereId"`
	Deleted   bool      `json:"deleted"`
	UpdatedAt time.Time `json:"updatedAt"`
}
