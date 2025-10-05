package core

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateMongoIndexes() []mongo.IndexModel {
	indexes := []mongo.IndexModel{
		// 1) Único por key (ya lo tenías)
		{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetName("uniq_key").SetUnique(true)},

		// 2) Búsquedas por namespace+entityId (ya lo tenías)
		{Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "entityId", Value: 1}}, Options: options.Index().SetName("ns_entity")},

		// 3) Watcher (pending -> headOk) (ya lo tenías)
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "nextCheckAt", Value: 1}}, Options: options.Index().SetName("status_nextCheckAt")},

		// 4) Listar por scope (solo ready)
		{
			Keys: bson.D{
				{Key: "scopeType", Value: 1},
				{Key: "scopeId", Value: 1},
				{Key: "profile", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().
				SetName("scope_ready_createdAt_desc").
				SetPartialFilterExpression(bson.M{"status": "ready"}),
		},

		// 5) Listar por uploader (solo ready)
		{
			Keys: bson.D{
				{Key: "uploaderId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().
				SetName("uploader_ready_createdAt_desc").
				SetPartialFilterExpression(bson.M{"status": "ready"}),
		},

		// 6) Por mediaType (solo ready) — útil para feeds/analíticas
		{
			Keys: bson.D{
				{Key: "mediaType", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().
				SetName("mediaType_ready_createdAt_desc").
				SetPartialFilterExpression(bson.M{"status": "ready"}),
		},

		// 7) TTL opcional: documentos con expireAt
		//    IMPORTANTE: TTL NO permite partial; usa un campo `expireAt` solo en docs a expirar
		{
			Keys: bson.D{{Key: "expireAt", Value: 1}},
			Options: options.Index().
				SetName("ttl_expireAt").
				SetExpireAfterSeconds(0), // 0 => a la hora exacta de expireAt
		},
	}

	return indexes
}
