package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
)

type TxHash struct {
	Hash        string  `bson:"hash"`
	BlockNumber *BigInt `bson:"blockNumber"`
	BlockHash   string  `bson:"blockHash"`
}

func (c *DbClient) GetTxHash(hash string) (*TxHash, error) {
	return c.findTxHashByHash(hash)
}

func (c *DbClient) SaveTxHash(hash string, blockNumber *big.Int, blockHash string) error {
	txHash, err := c.GetTxHash(hash)
	if err != nil {
		return err
	}
	if txHash == nil {
		txHash = &TxHash{
			Hash:        hash,
			BlockNumber: &BigInt{blockNumber},
			BlockHash:   blockHash,
		}
		return c.insertTxHash(txHash)
	}
	if txHash.BlockNumber.Equals2(blockNumber) && txHash.BlockHash == blockHash {
		return nil
	}
	txHash.BlockNumber = &BigInt{blockNumber}
	txHash.BlockHash = blockHash
	return c.updateTxHashByHash(hash, txHash)
}

func (c *DbClient) findTxHashByHash(hash string) (*TxHash, error) {
	var doc *TxHash
	err := c.Collection(COLLECTION_TX_HASHES).FindOne(
		context.TODO(),
		bson.D{{Key: "hash", Value: hash}},
	).Decode(&doc)
	if c.isEmptyResultError(err) {
		return nil, nil
	}
	return doc, err
}

func (c *DbClient) insertTxHash(txHash *TxHash) error {
	_, err := c.Collection(COLLECTION_TX_HASHES).InsertOne(
		context.TODO(),
		txHash,
	)
	return err
}

func (c *DbClient) updateTxHashByHash(hash string, txHash *TxHash) error {
	_, err := c.Collection(COLLECTION_TX_HASHES).UpdateOne(
		context.TODO(),
		bson.D{{Key: "hash", Value: hash}},
		bson.D{{Key: "$set", Value: txHash}},
	)
	return err
}
