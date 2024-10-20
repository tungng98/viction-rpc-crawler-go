package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlockHash struct {
	Hash           string  `bson:"hash"`
	BlockNumber    *BigInt `bson:"blockNumber"`
	BlockNumberHex string  `bson:"blockNumberHex"`
}

func NewBlockHash(hash string, blockNumber *big.Int) *BlockHash {
	return &BlockHash{
		Hash:        hash,
		BlockNumber: &BigInt{blockNumber},
	}
}

func (c *DbClient) GetBlockHash(hash string) (*BlockHash, error) {
	return c.findBlockHashByHash(hash)
}

func (c *DbClient) GetBlockHashes(hashes []string) ([]*BlockHash, error) {
	return c.findBlockHashesByHashes(hashes)
}

func (c *DbClient) SaveBlockHash(hash string, blockNumber *big.Int) error {
	blockHash, err := c.GetBlockHash(hash)
	if err != nil {
		return err
	}
	if blockHash == nil {
		blockHash = NewBlockHash(hash, blockNumber)
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

func (c *DbClient) SaveBlockHashes(newBlockHashes []*BlockHash, chnagedBlockHashes []*BlockHash) (*BulkWriteResult, error) {
	return c.writeBlockHashes(newBlockHashes, chnagedBlockHashes)
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

func (c *DbClient) findBlockHashesByHashes(hashes []string) ([]*BlockHash, error) {
	var docs []*BlockHash
	cursor, err := c.Collection(COLLECTION_BLOCK_HASHES).Find(
		context.TODO(),
		bson.D{{Key: "hash", Value: bson.D{{Key: "$in", Value: hashes}}}},
	)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &docs)
	return docs, err
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

func (c *DbClient) writeBlockHashes(newBlockHashes []*BlockHash, changedBlockHashes []*BlockHash) (*BulkWriteResult, error) {
	docs := []mongo.WriteModel{}
	for _, blockHash := range newBlockHashes {
		blockHash.BlockNumberHex = "0x" + blockHash.BlockNumber.Hex()
		docs = append(docs, &mongo.InsertOneModel{Document: blockHash})
	}
	for _, blockHash := range changedBlockHashes {
		blockHash.BlockNumberHex = "0x" + blockHash.BlockNumber.Hex()
		docs = append(docs, &mongo.UpdateOneModel{Filter: bson.D{{Key: "hash", Value: blockHash.Hash}}, Update: bson.D{{Key: "$set", Value: blockHash}}})
	}
	result, err := c.Collection(COLLECTION_BLOCK_HASHES).BulkWrite(
		context.TODO(),
		docs,
	)
	return newBulkWriteResult(result), err
}
