package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoValues = "values"
	mongoUsers  = "users"
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

func (m *MongoDb) Put(ctx context.Context, value *BValue) (err error) {
	_, err = m.db.Collection(mongoValues).InsertOne(ctx, value)
	return err
}

func (m *MongoDb) GetLast(ctx context.Context, userID int) (value int, err error) {
	opt := options.FindOne().SetSort(bson.D{primitive.E{Key: "timestamp", Value: -1}})
	result := m.db.Collection(mongoValues).FindOne(ctx, bson.D{primitive.E{Key: "userid", Value: userID}}, opt)
	if result == nil {
		return value, fmt.Errorf("failed to get last stored value from db")
	}

	bvalue := new(BValue)
	if err := result.Decode(bvalue); err != nil {
		return value, fmt.Errorf("failed to decode got value from db, %s", err)
	}
	value = bvalue.Value
	return
}

func (m *MongoDb) AddUser(ctx context.Context, user *TUser) (err error) {
	_, err = m.db.Collection(mongoUsers).InsertOne(ctx, user)
	return
}
