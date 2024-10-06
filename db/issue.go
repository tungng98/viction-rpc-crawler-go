package db

import (
	"context"
	"math/big"
	"time"
)

type Issue struct {
	Type        string     `bson:"type"`
	TxHash      string     `bson:"txHash"`
	BlockNumber *BigInt    `bson:"blockNumber"`
	BlockHash   string     `bson:"blockHash"`
	Timestamp   *Timestamp `bson:"timestamp"`
	TimeString  string     `bson:"timestring"`

	Extras map[string]interface{} `bson:"extras"`
}

func (c *DbClient) SaveDuplicatedTxHashIssue(txHash string, blockNumber *big.Int, blockHash string, prevBlockNumber *big.Int, prevBlockHash string) error {
	extras := map[string]interface{}{
		"prevBlockNumber": &BigInt{prevBlockNumber},
		"prevBlockHash":   prevBlockHash,
	}
	issue := &Issue{
		Type:        "duplicated_tx_hash",
		TxHash:      txHash,
		BlockNumber: &BigInt{blockNumber},
		BlockHash:   blockHash,
		Extras:      extras,
	}
	return c.insertIssue(issue)
}

func (c *DbClient) insertIssue(issue *Issue) error {
	now := time.Now()
	issue.Timestamp = &Timestamp{now}
	issue.TimeString = now.UTC().Format(time.RFC3339Nano)
	_, err := c.Collection(COLLECTION_ISSUES).InsertOne(
		context.TODO(),
		issue,
	)
	return err
}
