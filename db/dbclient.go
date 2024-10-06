package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION_BLOCKS    = "blocks"
	COLLECTION_TX_HASHES = "txHashes"
)

type DbClient struct {
	c  *mongo.Client
	db string
}

func Connect(uri string, database string) (*DbClient, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &DbClient{client, database}, nil
}

func (c *DbClient) Collection(collection string) *mongo.Collection {
	return c.c.Database(c.db).Collection(collection)
}

func (c *DbClient) Disconnect() {
	c.c.Disconnect(context.TODO())
}

func (c *DbClient) isEmptyResult(result *mongo.SingleResult) bool {
	return c.isEmptyResultError(result.Err())
}

func (c *DbClient) isEmptyResultError(err error) bool {
	return err != nil && err.Error() == "mongo: no documents in result"
}
