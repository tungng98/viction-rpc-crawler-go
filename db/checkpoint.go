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

func (c *DbClient) GetHighestBlock() (*Checkpoint, error) {
	return c.findBlockByType("highest")
}

func (c *DbClient) SaveHighestBlock(number *big.Int) error {
	checkpoint, err := c.GetHighestBlock()
	if err != nil {
		return err
	}
	if checkpoint == nil {
		checkpoint = &Checkpoint{
			Type:        "highest",
			BlockNumber: &BigInt{number},
		}
		return c.insertCheckpoint(checkpoint)
	}
	if checkpoint.BlockNumber.Equals2(number) {
		return nil
	}
	checkpoint.BlockNumber = &BigInt{number}
	return c.updateCheckpointByType("highest", checkpoint)
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

func (c *DbClient) insertCheckpoint(block *Checkpoint) error {
	block.BlockNumberHex = "0x" + block.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_CHECKPOINTS).InsertOne(
		context.TODO(),
		block,
	)
	return err
}

func (c *DbClient) updateCheckpointByType(typ string, block *Checkpoint) error {
	block.BlockNumberHex = "0x" + block.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_CHECKPOINTS).UpdateOne(
		context.TODO(),
		bson.D{{Key: "type", Value: typ}},
		bson.D{{Key: "$set", Value: block}},
	)
	return err
}
