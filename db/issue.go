package db

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"time"
)

const (
	ERROR_ISSUE uint16 = iota
	REORG_BLOCK_ISSUE
	DUPLICATED_TX_HASH_ISSUE
)

type Issue struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Type        uint16 `gorm:"column:type"`
	BlockNumber uint64 `gorm:"column:block_number"`
	BlockHash   []byte `gorm:"column:block_hash"`
	TxHash      []byte `gorm:"column:tx_hash"`
	Timestamp   int64  `gorm:"column:timestamp"`
	Status      bool   `gorm:"column:status"`
	Hash        []byte `gorm:"column:hash;unique"`

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

func (c *DbClient) NewErrorIssue(txHash []byte, blockNumber uint64, blockHash []byte, err error) *Issue {
	extras := map[string]interface{}{
		"error": err.Error(),
	}
	issue := &Issue{
		Type:        ERROR_ISSUE,
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		TxHash:      txHash,
		Extras:      extras,
	}
	return issue
}

func NewReorgBlockIssue(blockNumber uint64, blockHash, prevBlockHash []byte) *Issue {
	extras := map[string]interface{}{
		"prev_block_hash": prevBlockHash,
	}
	issue := &Issue{
		Type:        REORG_BLOCK_ISSUE,
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		TxHash:      make([]byte, 0),
		Extras:      extras,
	}
	return issue
}

func NewDuplicatedTxHashIssue(txHash []byte, blockNumber uint64, blockHash []byte, prevBlockNumber uint64, prevBlockHash []byte) *Issue {
	extras := map[string]interface{}{
		"prev_block_number": prevBlockNumber,
		"prev_block_hash":   prevBlockHash,
	}
	issue := &Issue{
		Type:        DUPLICATED_TX_HASH_ISSUE,
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		TxHash:      txHash,
		Extras:      extras,
	}
	return issue
}

func (c *DbClient) SaveErrorIssue(txHash []byte, blockNumber uint64, blockHash []byte, err error) error {
	issue := c.NewErrorIssue(txHash, blockNumber, blockHash, err)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveReorgBlockIssue(blockNumber uint64, blockHash, prevBlockHash []byte) error {
	issue := NewReorgBlockIssue(blockNumber, blockHash, prevBlockHash)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveDuplicatedTxHashIssue(txHash []byte, blockNumber uint64, blockHash []byte, prevBlockNumber uint64, prevBlockHash []byte) error {
	issue := NewDuplicatedTxHashIssue(txHash, blockNumber, blockHash, prevBlockNumber, prevBlockHash)
	return c.insertIssue(issue)
}

func (c *DbClient) SaveIssues(issues []*Issue) error {
	return c.writeIssues(issues)
}

func (c *DbClient) insertIssue(issue *Issue) error {
	now := time.Now().UnixMicro()
	issue.Checksum()
	issue.Timestamp = now
	result := c.d.Create(issue)
	return result.Error
}

func (c *DbClient) writeIssues(newIssues []*Issue) error {
	now := time.Now().UnixMicro()
	for _, issue := range newIssues {
		issue.Checksum()
		issue.Timestamp = now
	}
	result := c.d.CreateInBatches(newIssues, len(newIssues))
	return result.Error
}
