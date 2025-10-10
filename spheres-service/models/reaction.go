package models

import "go.mongodb.org/mongo-driver/v2/bson"

type SphereContentReaction struct {
	ID       bson.ObjectID  `bson:"_id" json:"_id"`
	SphereID string         `bson:"sphereId" json:"sphereId"`
	UserID   bson.ObjectID  `bson:"userId" json:"userId"`
	ParentID *bson.ObjectID `bson:"parentId,omitempty" json:"parentId"`
	Symbol   string         `bson:"symbol" json:"symbol"`
	Action   string         `bson:"-" json:"action"` // "added" o "removed"
}
