// mongo_conn.go — helper MongoDB (driver v2) con cliente singleton.
//
// Variables de entorno reconocidas:
//   MONGO_URI  → "mongodb://user:pass@host:27017"
//   MONGO_DB   → nombre de la base (por defecto "moonmap")
//
// Uso mínimo:
//   coll := MustGetCollection("projects")
//   coll.InsertOne(ctx, bson.M{"foo": "bar"})
//
// El cliente y el contexto viven como singleton dentro del paquete; solo hay
// un `Disconnect` al final del programa si llamas CloseMongo().

package persistence

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/helpers"
)

var (
	mongoURI      = ""
	mongoDatabase = ""

	clientOnce  sync.Once
	mongoClient *mongo.Client
)

// initClient abre la conexión una sola vez (lazy).
func InitMongo() {
	clientOnce.Do(func() {
		mongoURI = helpers.GetEnv("MONGO_URI", "mongodb://mongomongo:27017")
		mongoDatabase = helpers.GetEnv("MONGO_DB", "mongomongo")
		mongoTls := helpers.GetEnv("MONGO_TLS", "false") == "true"
		mongoDirectConn := helpers.GetEnv("MONGO_DIRECT_CONNECTION", "false") == "true"
		mongoReplicaSet := helpers.GetEnv("MONGO_REPLICA_SET", "rs0")

		opts := options.Client().ApplyURI(mongoURI).SetMinPoolSize(10).SetServerSelectionTimeout(30 * time.Second)
		opts.SetAuth(options.Credential{
			Username:      helpers.GetEnv("MONGO_USER", "mongomongo"),
			Password:      helpers.GetEnv("MONGO_PASS", "mongomongo"),
			AuthSource:    mongoDatabase,
			AuthMechanism: "SCRAM-SHA-256",
		})
		if mongoTls {
			opts.SetTLSConfig(&tls.Config{})
		}

		opts.SetDirect(mongoDirectConn)
		if !mongoDirectConn {
			opts.SetReplicaSet(mongoReplicaSet)
			logrus.Infof("setting mongo replicaset %v", mongoReplicaSet)
		}

		var err error
		logrus.WithField("mongoURI", mongoURI).Info("connecting to mongo")
		mongoClient, err = mongo.Connect(opts)
		if err != nil {
			logrus.WithError(err).Fatal("mongo connect")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logrus.Info("ping to mongo...")
		if err := mongoClient.Ping(ctx, nil); err != nil {
			_ = mongoClient.Disconnect(context.Background())
			logrus.WithError(err).Fatal("mongo ping")
		}
		logrus.Info("mongo ping success")
	})
}

// GetMongoClient devuelve el cliente y contexto global, inicializando si es necesario.
func GetMongoClient() *mongo.Client {
	InitMongo()
	return mongoClient
}

// MustGetCollection devuelve la colección deseada utilizando el cliente singleton.
func MustGetCollection(name string) *mongo.Collection {
	return GetMongoClient().Database(mongoDatabase).Collection(name)
}

// CloseMongo libera recursos (llámalo con defer en main si quieres un cierre limpio).
func CloseMongo() {
	logrus.Info("closing mongo pool connection")
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}
}

type idRange struct {
	Min bson.ObjectID
	Max bson.ObjectID
}

func ComputeIDRanges(ctx context.Context, col *mongo.Collection, partitions int) ([]idRange, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$bucketAuto", Value: bson.M{
			"groupBy": "$_id",
			"buckets": partitions,
			"output": bson.M{
				"min": bson.M{"$min": "$_id"},
				"max": bson.M{"$max": "$_id"},
			},
		}}},
	}
	cur, err := col.Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var ranges []idRange
	for cur.Next(ctx) {
		var m struct {
			Min bson.ObjectID `bson:"min"`
			Max bson.ObjectID `bson:"max"`
		}
		if err := cur.Decode(&m); err == nil {
			ranges = append(ranges, idRange{Min: m.Min, Max: m.Max})
		}
	}
	if len(ranges) == 0 && cur.Err() == nil {
		var first struct {
			ID bson.ObjectID `bson:"_id"`
		}
		var last struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err := col.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "_id", Value: 1}})).Decode(&first); err == nil {
			if err := col.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})).Decode(&last); err == nil {
				ranges = []idRange{{Min: first.ID, Max: last.ID}}
			}
		}
	}
	return ranges, cur.Err()
}
