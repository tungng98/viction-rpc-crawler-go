package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
)

type BlockHash struct {
	Hash           string  `bson:"hash"`
	BlockNumber    *BigInt `bson:"blockNumber"`
	BlockNumberHex string  `bson:"blockNumberHex"`
}

func (c *DbClient) GetBlockHash(hash string) (*BlockHash, error) {
	return c.findBlockHashByHash(hash)
}

func (c *DbClient) SaveBlockHash(hash string, blockNumber *big.Int) error {
	blockHash, err := c.GetBlockHash(hash)
	if err != nil {
		return err
	}
	if blockHash == nil {
		blockHash = &BlockHash{
			Hash:        hash,
			BlockNumber: &BigInt{blockNumber},
		}
		return c.insertBlockHash(blockHash)
	}
	if blockHash.BlockNumber.Equals2(blockNumber) {
		return nil
	}
	err = c.SaveDuplicatedBlockHashIssue(hash, blockNumber, blockHash.BlockNumber.N)
	if err != nil {
		return err
	}
	blockHash.BlockNumber = &BigInt{blockNumber}
	return c.updateBlockHashByHash(hash, blockHash)
}

func (c *DbClient) findBlockHashByHash(hash string) (*BlockHash, error) {
	var doc *BlockHash
	err := c.Collection(COLLECTION_BLOCK_HASHES).FindOne(
		context.TODO(),
		bson.D{{Key: "hash", Value: hash}},
	).Decode(&doc)
	if c.isEmptyResultError(err) {
		return nil, nil
	}
	return doc, err
}

func (c *DbClient) insertBlockHash(blockHash *BlockHash) error {
	blockHash.BlockNumberHex = "0x" + blockHash.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_BLOCK_HASHES).InsertOne(
		context.TODO(),
		blockHash,
	)
	return err
}

func (c *DbClient) updateBlockHashByHash(hash string, blockHash *BlockHash) error {
	blockHash.BlockNumberHex = "0x" + blockHash.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_BLOCK_HASHES).UpdateOne(
		context.TODO(),
		bson.D{{Key: "hash", Value: hash}},
		bson.D{{Key: "$set", Value: blockHash}},
	)
	return err
}
