package db

import (
	"math/big"
	"viction-rpc-crawler-go/x/ethutil"

	"github.com/tee8z/nullable"
)

type Block struct {
	ID                     uint64          `gorm:"column:id;primaryKey"`
	Hash                   []byte          `gorm:"column:hash;uniqueIndex"`
	ParentHash             []byte          `gorm:"column:parent_hash"`
	StateRoot              []byte          `gorm:"column:state_root"`
	TransactionsRoot       []byte          `gorm:"column:transaction_root"`
	ReceiptsRoot           []byte          `gorm:"column:receipts_root"`
	Timestamp              int64           `gorm:"column:timestamp"`
	Size                   uint16          `gorm:"column:size"`
	GasLimit               uint64          `gorm:"column:gas_limit"`
	GasUsed                uint64          `gorm:"column:gas_used"`
	TotalDifficulty        uint64          `gorm:"column:total_difficulty"`
	TransactionCount       uint16          `gorm:"column:transaction_count"`
	TransactionCountSystem nullable.Uint16 `gorm:"column:transaction_count_system"`
	TransactionCountDebug  nullable.Uint16 `gorm:"column:transaction_count_debug"`
	BlockMintDuration      nullable.Uint64 `gorm:"column:block_mint_duration"`
}

func NewBlock(blockNumber *big.Int, blockHash, parentHash []byte, stateRoot, transactionRoot, receiptsRoot []byte, timestamp int64, size uint16, gasLimit, gasUsed uint64, totalDifficult uint64,
	transactionCount uint16, transactionCountSystem, transactionCountDebug nullable.Uint16, blockMintDuration nullable.Uint64) *Block {
	return &Block{
		ID:                     blockNumber.Uint64(),
		Hash:                   blockHash,
		ParentHash:             parentHash,
		StateRoot:              stateRoot,
		TransactionsRoot:       transactionRoot,
		ReceiptsRoot:           receiptsRoot,
		Timestamp:              timestamp,
		Size:                   size,
		GasLimit:               gasLimit,
		GasUsed:                gasUsed,
		TotalDifficulty:        totalDifficult,
		TransactionCount:       transactionCount,
		TransactionCountSystem: transactionCountSystem,
		TransactionCountDebug:  transactionCountDebug,
		BlockMintDuration:      blockMintDuration,
	}
}

func (c *DbClient) GetBlock(id uint64) (*Block, error) {
	return c.findBlock(id)
}

func (c *DbClient) GetBlockByHash(hash []byte) (*Block, error) {
	return c.findBlockByHash(hash)
}

func (c *DbClient) GetBlocksByHashes(hashes [][]byte) ([]*Block, error) {
	return c.findBlocksByHashes(hashes)
}

func (c *DbClient) SaveBlock(newBlock *Block) error {
	block, err := c.findBlock(newBlock.ID)
	if err != nil {
		return err
	}
	if block == nil {
		return c.insertBlock(newBlock)
	}
	if !ethutil.BytesEqual(block.Hash, newBlock.Hash) {
		err = c.SaveReorgBlockIssue(newBlock.ID, newBlock.Hash, block.Hash)
		if err != nil {
			return err
		}
	}
	return c.updateBlockByID(block.ID, block)
}

func (c *DbClient) SaveBlocks(newBlocks []*Block, chnagedBlocks []*Block) error {
	return c.writeBlocks(newBlocks, chnagedBlocks)
}

func (c *DbClient) findBlock(id uint64) (*Block, error) {
	var doc *Block
	result := c.d.Model(&Block{}).
		Where("id = ?", id).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (c *DbClient) findBlockByHash(hash []byte) (*Block, error) {
	var doc *Block
	result := c.d.Model(&Block{}).
		Where("hash = ?", hash).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (c *DbClient) findBlocksByHashes(hashes [][]byte) ([]*Block, error) {
	var docs []*Block
	result := c.d.Model(&Block{}).
		Where("hash IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

func (c *DbClient) insertBlock(block *Block) error {
	result := c.d.Create(block)
	return result.Error
}

func (c *DbClient) updateBlockByID(id uint64, block *Block) error {
	result := c.d.Model(&Block{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"hash":                     block.Hash,
			"parent_hash":              block.ParentHash,
			"state_root":               block.StateRoot,
			"transaction_root":         block.TransactionsRoot,
			"receipts_root":            block.ReceiptsRoot,
			"timestamp":                block.Timestamp,
			"size":                     block.Size,
			"gas_limit":                block.GasLimit,
			"gas_used":                 block.GasUsed,
			"total_difficulty":         block.TotalDifficulty,
			"transaction_count":        block.TransactionCount,
			"transaction_count_system": block.TransactionCountSystem,
			"transaction_count_debug":  block.TransactionCountDebug,
			"block_mint_duration":      block.BlockMintDuration,
		})
	return result.Error
}

func (c *DbClient) writeBlocks(newBlocks []*Block, changedBlocks []*Block) error {
	tx := c.d.Begin()
	for _, block := range newBlocks {
		result := tx.Create(block)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	for _, block := range changedBlocks {
		result := tx.Model(&Block{}).
			Where("id = ?", block.ID).
			Updates(map[string]interface{}{
				"hash":                     block.Hash,
				"parent_hash":              block.ParentHash,
				"state_root":               block.StateRoot,
				"transaction_root":         block.TransactionsRoot,
				"receipts_root":            block.ReceiptsRoot,
				"timestamp":                block.Timestamp,
				"size":                     block.Size,
				"gas_limit":                block.GasLimit,
				"gas_used":                 block.GasUsed,
				"total_difficulty":         block.TotalDifficulty,
				"transaction_count":        block.TransactionCount,
				"transaction_count_system": block.TransactionCountSystem,
				"transaction_count_debug":  block.TransactionCountDebug,
				"block_mint_duration":      block.BlockMintDuration,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}
