package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
)

type Checkpoint struct {
	Type           string  `bson:"type"`
	BlockNumber    *BigInt `bson:"blockNumber"`
	BlockNumberHex string  `bson:"blockNumberHex"`
}

func (c *DbClient) GetHighestIndexBlock() (*Checkpoint, error) {
	return c.findBlockByType("highest_index")
}

func (c *DbClient) SaveHighestIndexBlock(number *big.Int) error {
	checkpoint, err := c.GetHighestIndexBlock()
	if err != nil {
		return err
	}
	if checkpoint == nil {
		checkpoint = &Checkpoint{
			Type:        "highest_index",
			BlockNumber: &BigInt{number},
		}
		return c.insertCheckpoint(checkpoint)
	}
	if checkpoint.BlockNumber.Equals2(number) {
		return nil
	}
	checkpoint.BlockNumber = &BigInt{number}
	return c.updateCheckpointByType("highest_index", checkpoint)
}

func (c *DbClient) findBlockByType(typ string) (*Checkpoint, error) {
	var doc *Checkpoint
	err := c.Collection(COLLECTION_CHECKPOINTS).FindOne(
		context.TODO(),
		bson.D{{Key: "type", Value: typ}},
	).Decode(&doc)
	if c.isEmptyResultError(err) {
		return nil, nil
	}
	return doc, err
}

func (c *DbClient) insertCheckpoint(checkpoint *Checkpoint) error {
	checkpoint.BlockNumberHex = "0x" + checkpoint.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_CHECKPOINTS).InsertOne(
		context.TODO(),
		checkpoint,
	)
	return err
}

func (c *DbClient) updateCheckpointByType(typ string, checkpoint *Checkpoint) error {
	checkpoint.BlockNumberHex = "0x" + checkpoint.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_CHECKPOINTS).UpdateOne(
		context.TODO(),
		bson.D{{Key: "type", Value: typ}},
		bson.D{{Key: "$set", Value: checkpoint}},
	)
	return err
}
