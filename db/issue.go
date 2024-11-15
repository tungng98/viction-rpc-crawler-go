package db

import (
	"context"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Issue struct {
	Type           string     `bson:"type"`
	TxHash         string     `bson:"txHash"`
	BlockNumber    *BigInt    `bson:"blockNumber"`
	BlockNumberHex string     `bson:"blockNumberHex"`
	BlockHash      string     `bson:"blockHash"`
	Timestamp      *Timestamp `bson:"timestamp"`
	TimeString     string     `bson:"timestring"`

	Extras map[string]interface{} `bson:"extras"`
}

func NewDuplicatedBlockHashIssue(blockHash string, blockNumber *big.Int, prevBlockNumber *big.Int) *Issue {
	extras := map[string]interface{}{
		"prevBlockNumber":    &BigInt{prevBlockNumber},
		"prevBlockNumberHex": "0x" + (&BigInt{prevBlockNumber}).Hex(),
	}
	issue := &Issue{
		Type:        "duplicated_block_hash",
		TxHash:      "",
		BlockNumber: &BigInt{blockNumber},
		BlockHash:   blockHash,
		Extras:      extras,
	}
	return issue
}

func NewDuplicatedTxHashIssue(txHash string, blockNumber *big.Int, blockHash string, prevBlockNumber *big.Int, prevBlockHash string) *Issue {
	extras := map[string]interface{}{
		"prevBlockNumber":    &BigInt{prevBlockNumber},
		"prevBlockNumberHex": "0x" + (&BigInt{prevBlockNumber}).Hex(),
		"prevBlockHash":      prevBlockHash,
	}
	issue := &Issue{
		Type:        "duplicated_tx_hash",
		TxHash:      txHash,
		BlockNumber: &BigInt{blockNumber},
		BlockHash:   blockHash,
		Extras:      extras,
	}
	return issue
}

func (c *DbClient) NewErrorIssue(txHash string, blockHash string, blockNumber *big.Int, err error) *Issue {
	extras := map[string]interface{}{
		"error": err.Error(),
	}
	issue := &Issue{
		Type:        "error",
		TxHash:      txHash,
		BlockNumber: &BigInt{blockNumber},
		BlockHash:   blockHash,
		Extras:      extras,
	}
	return issue
}

func (c *DbClient) SaveDuplicatedBlockHashIssue(blockHash string, blockNumber *big.Int, prevBlockNumber *big.Int) error {
	issue := NewDuplicatedBlockHashIssue(blockHash, blockNumber, prevBlockNumber)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveDuplicatedTxHashIssue(txHash string, blockNumber *big.Int, blockHash string, prevBlockNumber *big.Int, prevBlockHash string) error {
	issue := NewDuplicatedTxHashIssue(txHash, blockNumber, blockHash, prevBlockNumber, prevBlockHash)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveErrorIssue(txHash string, blockHash string, blockNumber *big.Int, err error) error {
	issue := c.NewErrorIssue(txHash, blockHash, blockNumber, err)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveIssues(issues []*Issue) (*BulkWriteResult, error) {
	return c.writeIssues(issues)
}

func (c *DbClient) insertIssue(issue *Issue) error {
	now := time.Now()
	issue.BlockNumberHex = "0x" + issue.BlockNumber.Hex()
	issue.Timestamp = &Timestamp{now}
	issue.TimeString = now.UTC().Format(time.RFC3339Nano)
	_, err := c.Collection(COLLECTION_ISSUES).InsertOne(
		context.TODO(),
		issue,
	)
	return err
}

func (c *DbClient) writeIssues(newIssues []*Issue) (*BulkWriteResult, error) {
	now := time.Now()
	docs := []mongo.WriteModel{}
	for _, issue := range newIssues {
		issue.BlockNumberHex = "0x" + issue.BlockNumber.Hex()
		issue.Timestamp = &Timestamp{now}
		issue.TimeString = now.UTC().Format(time.RFC3339Nano)
		docs = append(docs, &mongo.InsertOneModel{Document: issue})
	}
	result, err := c.Collection(COLLECTION_ISSUES).BulkWrite(
		context.TODO(),
		docs,
	)
	return newBulkWriteResult(result), err
}
