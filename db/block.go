package db

import (
	"math/big"

	"github.com/gurukami/typ"
	"github.com/shopspring/decimal"
)

type Block struct {
	ID                     uint64          `gorm:"column:id;primaryKey"`
	Hash                   string          `gorm:"column:hash;length:32;uniqueIndex"`
	ParentHash             string          `gorm:"column:parent_hash;length:32"`
	Timestamp              int64           `gorm:"column:timestamp"`
	Size                   uint16          `gorm:"column:size"`
	GasLimit               uint64          `gorm:"column:gas_limit"`
	GasUsed                uint64          `gorm:"column:gas_used"`
	Difficulty             decimal.Decimal `gorm:"column:difficulty;type:decimal(78,0)"`
	TotalDifficulty        decimal.Decimal `gorm:"column:total_difficulty;type:decimal(78,0)"`
	TransactionCount       typ.NullUint16  `gorm:"column:transaction_count"`
	TransactionCountSystem typ.NullUint16  `gorm:"column:transaction_count_system"`
	TransactionCountDebug  typ.NullUint16  `gorm:"column:transaction_count_debug"`
	BlockMintDuration      typ.NullUint64  `gorm:"column:block_mint_duration"`
	UncleHash              []byte          `gorm:"column:uncle_hash;length:32"`
	StateRoot              []byte          `gorm:"column:state_root;length:32"`
	TransactionsRoot       []byte          `gorm:"column:transaction_root;length:32"`
	ReceiptsRoot           []byte          `gorm:"column:receipts_root;length:32"`
	LogsBloom              []byte          `gorm:"column:logs_bloom;length:256"`
	Miner                  []byte          `gorm:"column:miner;length:20"`
	ExtraData              []byte          `gorm:"column:extra_data"`
	MixDigest              []byte          `gorm:"column:mix_digest"`
	Nonce                  []byte          `gorm:"column:nonce"`
	Validator              []byte          `gorm:"column:validator"`
	Creator                typ.NullString  `gorm:"column:creator"`
	Attestor               typ.NullString  `gorm:"column:attestor"`
}

func NewBlock(blockNumber *big.Int, blockHash, parentHash string, timestamp int64, size uint16, gasLimit, gasUsed uint64, difficulty, totalDifficulty *big.Int,
	transactionCount, transactionCountSystem, transactionCountDebug typ.NullUint16, blockMintDuration typ.NullUint64,
	uncleHash []byte, stateRoot, transactionsRoot, receiptsRoot, logsBloom []byte,
	miner, extraData, mixDigest, nonce, validator []byte, creator, attestor typ.NullString) *Block {
	return &Block{
		ID:                     blockNumber.Uint64(),
		Hash:                   blockHash,
		ParentHash:             parentHash,
		Timestamp:              timestamp,
		Size:                   size,
		GasLimit:               gasLimit,
		GasUsed:                gasUsed,
		Difficulty:             decimal.NewFromBigInt(difficulty, 0),
		TotalDifficulty:        decimal.NewFromBigInt(totalDifficulty, 0),
		TransactionCount:       transactionCount,
		TransactionCountSystem: transactionCountSystem,
		TransactionCountDebug:  transactionCountDebug,
		BlockMintDuration:      blockMintDuration,
		UncleHash:              uncleHash,
		StateRoot:              stateRoot,
		TransactionsRoot:       transactionsRoot,
		ReceiptsRoot:           receiptsRoot,
		LogsBloom:              logsBloom,
		Miner:                  miner,
		ExtraData:              extraData,
		MixDigest:              mixDigest,
		Nonce:                  nonce,
		Validator:              validator,
		Creator:                creator,
		Attestor:               attestor,
	}
}

func (c *DbClient) GetBlock(id uint64) (*Block, error) {
	return c.findBlock(id)
}

func (c *DbClient) GetBlocks(ids []uint64) ([]*Block, error) {
	return c.findBlocks(ids)
}

func (c *DbClient) GetBlockByHash(hash string) (*Block, error) {
	return c.findBlockByHash(hash)
}

func (c *DbClient) GetBlocksByHashes(hashes []string) ([]*Block, error) {
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
	if block.Hash != newBlock.Hash {
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

func (c *DbClient) findBlocks(ids []uint64) ([]*Block, error) {
	var docs []*Block
	result := c.d.Model(&Block{}).
		Where("id IN ?", ids).
		Find(&docs)
	return docs, result.Error
}

func (c *DbClient) findBlockByHash(hash string) (*Block, error) {
	var doc *Block
	result := c.d.Model(&Block{}).
		Where("hash = ?", hash).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (c *DbClient) findBlocksByHashes(hashes []string) ([]*Block, error) {
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
			"timestamp":                block.Timestamp,
			"size":                     block.Size,
			"gas_limit":                block.GasLimit,
			"gas_used":                 block.GasUsed,
			"difficulty":               block.Difficulty,
			"total_difficulty":         block.TotalDifficulty,
			"transaction_count":        block.TransactionCount,
			"transaction_count_system": block.TransactionCountSystem,
			"transaction_count_debug":  block.TransactionCountDebug,
			"block_mint_duration":      block.BlockMintDuration,
			"uncle_hash":               block.UncleHash,
			"state_root":               block.StateRoot,
			"transaction_root":         block.TransactionsRoot,
			"receipts_root":            block.ReceiptsRoot,
			"logs_bloom":               block.LogsBloom,
			"miner":                    block.Miner,
			"extra_data":               block.ExtraData,
			"mix_digest":               block.MixDigest,
			"nonce":                    block.Nonce,
			"validator":                block.Validator,
			"creator":                  block.Creator,
			"attestor":                 block.Attestor,
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
				"timestamp":                block.Timestamp,
				"size":                     block.Size,
				"gas_limit":                block.GasLimit,
				"gas_used":                 block.GasUsed,
				"difficulty":               block.Difficulty,
				"total_difficulty":         block.TotalDifficulty,
				"transaction_count":        block.TransactionCount,
				"transaction_count_system": block.TransactionCountSystem,
				"transaction_count_debug":  block.TransactionCountDebug,
				"block_mint_duration":      block.BlockMintDuration,
				"uncle_hash":               block.UncleHash,
				"state_root":               block.StateRoot,
				"transaction_root":         block.TransactionsRoot,
				"receipts_root":            block.ReceiptsRoot,
				"logs_bloom":               block.LogsBloom,
				"miner":                    block.Miner,
				"extra_data":               block.ExtraData,
				"mix_digest":               block.MixDigest,
				"nonce":                    block.Nonce,
				"validator":                block.Validator,
				"creator":                  block.Creator,
				"attestor":                 block.Attestor,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}
