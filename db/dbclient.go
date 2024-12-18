package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION_BLOCK_HASHES = "blockHashes"
	COLLECTION_CHECKPOINTS  = "checkpoints"
	COLLECTION_ISSUES       = "issues"
	COLLECTION_TX_HASHES    = "txHashes"
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

func (c *DbClient) isEmptyResultError(err error) bool {
	return err != nil && err.Error() == "mongo: no documents in result"
}

type BulkWriteResult struct {
	InsertedCount int64
	MatchedCount  int64
	ModifiedCount int64
	DeletedCount  int64
	UpsertedCount int64
	UpsertedIDs   map[int64]interface{}
}

func newBulkWriteResult(result *mongo.BulkWriteResult) *BulkWriteResult {
	if result == nil {
		return nil
	}
	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		DeletedCount:  result.DeletedCount,
	}
}
