package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDb struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(mgoURL string, mgoDB string) (store Store, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mgoURL))
	if err != nil {
		return
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return
	}

	store = &MongoDb{client: client, db: client.Database(mgoDB)}
	return store, err
}

func (m *MongoDb) Put(context.Context) error {
	return nil
}
