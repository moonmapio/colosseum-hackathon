package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	collEvents   *mongo.Collection
	collSessions *mongo.Collection
	collWallets  *mongo.Collection
}

func NewStore(events, sessions, wallets *mongo.Collection) *Store {
	return &Store{collEvents: events, collSessions: sessions, collWallets: wallets}
}

// Guardar evento crudo
func (s *Store) SaveEvent(ctx context.Context, evt map[string]any) {
	doc := bson.M{"evt": evt, "at": time.Now().UTC()}
	_, _ = s.collEvents.InsertOne(ctx, doc)
}

// Agregar participante (join)
func (s *Store) AddParticipant(ctx context.Context, sphereId, userId string, evtTime time.Time) {
	_, _ = s.collSessions.UpdateOne(ctx,
		bson.M{
			"sphereId": sphereId,
			"$or": []bson.M{
				{"lastUpdated": bson.M{"$lt": evtTime}},
				{"lastUpdated": bson.M{"$exists": false}},
			},
		},
		bson.M{
			"$setOnInsert": bson.M{"startedAt": evtTime, "sphereId": sphereId},
			"$set":         bson.M{"active": true, "lastUpdated": evtTime},
			"$addToSet":    bson.M{"participants": userId},
		},
		options.UpdateOne().SetUpsert(true),
	)
}

// Eliminar participante (leave)
func (s *Store) RemoveParticipant(ctx context.Context, sphereId, userId string, evtTime time.Time) {
	_, _ = s.collSessions.UpdateOne(ctx,
		bson.M{
			"sphereId":    sphereId,
			"lastUpdated": bson.M{"$lt": evtTime},
		},
		bson.M{
			"$pull": bson.M{"participants": userId},
			"$set":  bson.M{"lastUpdated": evtTime},
		},
	)

	// Verificar si quedó vacío
	var doc struct {
		Participants []string `bson:"participants"`
	}
	if err := s.collSessions.FindOne(ctx, bson.M{"sphereId": sphereId}).Decode(&doc); err == nil {
		if len(doc.Participants) == 0 {
			_, _ = s.collSessions.UpdateOne(ctx,
				bson.M{"sphereId": sphereId},
				bson.M{"$set": bson.M{
					"active":      false,
					"endedAt":     evtTime,
					"lastUpdated": evtTime,
				}},
			)
		}
	}
}

// Cerrar sala (finished)
func (s *Store) CloseRoom(ctx context.Context, sphereId string, evtTime time.Time) {
	_, _ = s.collSessions.UpdateOne(ctx,
		bson.M{
			"sphereId": sphereId,
			"$or": []bson.M{
				{"lastUpdated": bson.M{"$lt": evtTime}},
				{"lastUpdated": bson.M{"$exists": false}},
			},
		},
		bson.M{"$set": bson.M{
			"active":       false,
			"participants": []string{},
			"endedAt":      evtTime,
			"lastUpdated":  evtTime,
		}},
	)
}

// Obtener stats con pipeline
func (s *Store) GetStats(ctx context.Context, sphereId string) (map[string]any, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "sphereId", Value: sphereId},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "wallets"},
			{Key: "localField", Value: "participants"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "users"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "sphereId", Value: 1},
			{Key: "active", Value: 1},
			{Key: "startedAt", Value: 1},
			{Key: "endedAt", Value: 1},
			{Key: "users.id", Value: "$users._id"},
			{Key: "users.name", Value: "$users.fullName"},
			{Key: "users.avatarUrl", Value: "$users.avatarUrl"},
			// todo add verified

			{Key: "users.walletAddress", Value: "$users.walletAddress"},
		}}},
	}

	cur, err := s.collSessions.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []map[string]any
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return out[0], nil
}
