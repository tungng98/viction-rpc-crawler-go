package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
)

type Block struct {
	Type   string  `bson:"type"`
	Number *BigInt `bson:"number"`
}

func (c *DbClient) GetHighestBlock() (*Block, error) {
	return c.findBlockByType("highest")
}

func (c *DbClient) SaveHighestBlock(number *big.Int) error {
	block, err := c.GetHighestBlock()
	if err != nil {
		return err
	}
	if block == nil {
		block = &Block{
			Type:   "highest",
			Number: &BigInt{number},
		}
		return c.insertBlock(block)
	}
	if block.Number.Equals2(number) {
		return nil
	}
	block.Number = &BigInt{number}
	return c.updateBlockByType("highest", block)
}

func (c *DbClient) findBlockByType(typ string) (*Block, error) {
	var doc *Block
	err := c.Collection(COLLECTION_BLOCKS).FindOne(
		context.TODO(),
		bson.D{{Key: "type", Value: typ}},
	).Decode(&doc)
	if c.isEmptyResultError(err) {
		return nil, nil
	}
	return doc, err
}

func (c *DbClient) insertBlock(block *Block) error {
	_, err := c.Collection(COLLECTION_BLOCKS).InsertOne(
		context.TODO(),
		block,
	)
	return err
}

func (c *DbClient) updateBlockByType(typ string, block *Block) error {
	_, err := c.Collection(COLLECTION_BLOCKS).UpdateOne(
		context.TODO(),
		bson.D{{Key: "type", Value: typ}},
		bson.D{{Key: "$set", Value: block}},
	)
	return err
}
