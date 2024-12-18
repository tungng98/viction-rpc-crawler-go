package db

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"time"

	"viction-rpc-crawler-go/x/ethutil"
)

const (
	ERROR_ISSUE uint16 = iota
	DUPLICATED_BLOCK_HASH_ISSUE
	DUPLICATED_TX_HASH_ISSUE
)

type Issue struct {
	ID        uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Type      uint16 `gorm:"column:type"`
	BlockHash []byte `gorm:"column:block_hash"`
	TxHash    []byte `gorm:"column:tx_hash"`
	Timestamp int64  `gorm:"column:timestamp"`
	Status    bool   `gorm:"column:status"`
	Hash      []byte `gorm:"column:hash;unique"`

	Extras map[string]interface{} `gorm:"column:extras;serializer:json"`
}

func (i *Issue) Checksum() {
	typeBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(typeBytes, i.Type)
	extraBytes, _ := json.Marshal(i.Extras)
	issueBytes := typeBytes
	issueBytes = append(issueBytes, i.BlockHash...)
	issueBytes = append(issueBytes, i.TxHash...)
	issueBytes = append(issueBytes, extraBytes...)
	hashBytes := sha256.Sum256(issueBytes)
	i.Hash = hashBytes[:]
}

func NewDuplicatedBlockHashIssue(blockHash string, blockNumber *big.Int, prevBlockNumber *big.Int) *Issue {
	extras := map[string]interface{}{
		"prev_block_number": prevBlockNumber.Uint64(),
	}
	issue := &Issue{
		Type:      DUPLICATED_BLOCK_HASH_ISSUE,
		TxHash:    ethutil.HexToBytes(""),
		BlockHash: ethutil.HexToBytes(blockHash),
		Extras:    extras,
	}
	return issue
}

func NewDuplicatedTxHashIssue(txHash string, blockNumber *big.Int, blockHash string, prevBlockNumber *big.Int, prevBlockHash string) *Issue {
	extras := map[string]interface{}{
		"prev_block_number": prevBlockNumber.Uint64(),
	}
	issue := &Issue{
		Type:      DUPLICATED_TX_HASH_ISSUE,
		TxHash:    ethutil.HexToBytes(txHash),
		BlockHash: ethutil.HexToBytes(blockHash),
		Extras:    extras,
	}
	return issue
}

func (c *DbClient) NewErrorIssue(txHash string, blockHash string, blockNumber *big.Int, err error) *Issue {
	extras := map[string]interface{}{
		"error": err.Error(),
	}
	issue := &Issue{
		Type:      ERROR_ISSUE,
		TxHash:    ethutil.HexToBytes(txHash),
		BlockHash: ethutil.HexToBytes(blockHash),
		Extras:    extras,
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
	now := time.Now().UnixMicro()
	issue.Checksum()
	issue.Timestamp = now
	result := c.d.Create(issue)
	return result.Error
}

func (c *DbClient) writeIssues(newIssues []*Issue) (*BulkWriteResult, error) {
	now := time.Now().UnixMicro()
	for _, issue := range newIssues {
		issue.Checksum()
		issue.Timestamp = now
	}
	result := c.d.CreateInBatches(newIssues, len(newIssues))
	return &BulkWriteResult{InsertedCount: int64(len(newIssues))}, result.Error
}
