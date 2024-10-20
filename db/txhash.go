package db

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TxHash struct {
	Hash           string  `bson:"hash"`
	BlockNumber    *BigInt `bson:"blockNumber"`
	BlockNumberHex string  `bson:"blockNumberHex"`
	BlockHash      string  `bson:"blockHash"`
}

func NewTxHash(hash string, blockNumber *big.Int, blockHash string) *TxHash {
	return &TxHash{
		Hash:        hash,
		BlockNumber: &BigInt{blockNumber},
		BlockHash:   blockHash,
	}
}

func (c *DbClient) GetTxHash(hash string) (*TxHash, error) {
	return c.findTxHashByHash(hash)
}

func (c *DbClient) GetTxHashes(hashes []string) ([]*TxHash, error) {
	return c.findTxHashesByHashes(hashes)
}

func (c *DbClient) SaveTxHash(hash string, blockNumber *big.Int, blockHash string) error {
	txHash, err := c.GetTxHash(hash)
	if err != nil {
		return err
	}
	if txHash == nil {
		txHash = NewTxHash(hash, blockNumber, blockHash)
		return c.insertTxHash(txHash)
	}
	if txHash.BlockNumber.Equals2(blockNumber) && txHash.BlockHash == blockHash {
		return nil
	}
	err = c.SaveDuplicatedTxHashIssue(hash, blockNumber, blockHash, txHash.BlockNumber.N, txHash.BlockHash)
	if err != nil {
		return err
	}
	txHash.BlockNumber = &BigInt{blockNumber}
	txHash.BlockHash = blockHash
	return c.updateTxHashByHash(hash, txHash)
}

func (c *DbClient) SaveTxHashes(newTxHashes []*TxHash, changedTxHashes []*TxHash) (*BulkWriteResult, error) {
	return c.writeTxHashes(newTxHashes, changedTxHashes)
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

func (c *DbClient) findTxHashesByHashes(hashes []string) ([]*TxHash, error) {
	var docs []*TxHash
	cursor, err := c.Collection(COLLECTION_TX_HASHES).Find(
		context.TODO(),
		bson.D{{Key: "hash", Value: bson.D{{Key: "$in", Value: hashes}}}},
	)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &docs)
	return docs, err
}

func (c *DbClient) insertTxHash(txHash *TxHash) error {
	txHash.BlockNumberHex = "0x" + txHash.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_TX_HASHES).InsertOne(
		context.TODO(),
		txHash,
	)
	return err
}

func (c *DbClient) updateTxHashByHash(hash string, txHash *TxHash) error {
	txHash.BlockNumberHex = "0x" + txHash.BlockNumber.Hex()
	_, err := c.Collection(COLLECTION_TX_HASHES).UpdateOne(
		context.TODO(),
		bson.D{{Key: "hash", Value: hash}},
		bson.D{{Key: "$set", Value: txHash}},
	)
	return err
}

func (c *DbClient) writeTxHashes(newTxHashes []*TxHash, changedTxHashes []*TxHash) (*BulkWriteResult, error) {
	docs := []mongo.WriteModel{}
	for _, txHash := range newTxHashes {
		txHash.BlockNumberHex = "0x" + txHash.BlockNumber.Hex()
		docs = append(docs, &mongo.InsertOneModel{Document: txHash})
	}
	for _, txHash := range changedTxHashes {
		txHash.BlockNumberHex = "0x" + txHash.BlockNumber.Hex()
		docs = append(docs, &mongo.UpdateOneModel{Filter: bson.D{{Key: "hash", Value: txHash.Hash}}, Update: bson.D{{Key: "$set", Value: txHash}}})
	}
	result, err := c.Collection(COLLECTION_TX_HASHES).BulkWrite(
		context.TODO(),
		docs,
	)
	return newBulkWriteResult(result), err
}
