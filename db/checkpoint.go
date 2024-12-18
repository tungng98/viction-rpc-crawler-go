package db

import (
	"math/big"
)

const (
	INDEX_CHECKPOINT = iota
	TRACE_CHECKPOINT
)

type Checkpoint struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Type        uint16 `gorm:"column:type"`
	BlockNumber uint64 `gorm:"column:block_number"`
}

func (c *DbClient) GetHighestIndexBlock() (*Checkpoint, error) {
	return c.findBlockByType(INDEX_CHECKPOINT)
}

func (c *DbClient) GetHighestTraceBlock() (*Checkpoint, error) {
	return c.findBlockByType(TRACE_CHECKPOINT)
}

func (c *DbClient) SaveHighestIndexBlock(number *big.Int) error {
	checkpoint, err := c.GetHighestIndexBlock()
	if err != nil {
		return err
	}
	if checkpoint == nil {
		checkpoint = &Checkpoint{
			Type:        INDEX_CHECKPOINT,
			BlockNumber: number.Uint64(),
		}
		return c.insertCheckpoint(checkpoint)
	}
	if checkpoint.BlockNumber == number.Uint64() {
		return nil
	}
	checkpoint.BlockNumber = number.Uint64()
	return c.updateCheckpointByType(INDEX_CHECKPOINT, checkpoint)
}

func (c *DbClient) SaveHighestTraceBlock(number *big.Int) error {
	checkpoint, err := c.GetHighestTraceBlock()
	if err != nil {
		return err
	}
	if checkpoint == nil {
		checkpoint = &Checkpoint{
			Type:        TRACE_CHECKPOINT,
			BlockNumber: number.Uint64(),
		}
		return c.insertCheckpoint(checkpoint)
	}
	if checkpoint.BlockNumber == number.Uint64() {
		return nil
	}
	checkpoint.BlockNumber = number.Uint64()
	return c.updateCheckpointByType(TRACE_CHECKPOINT, checkpoint)
}

func (c *DbClient) findBlockByType(typ uint16) (*Checkpoint, error) {
	var doc *Checkpoint
	result := c.d.Model(&Checkpoint{}).Where("type = ?", typ).First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (c *DbClient) insertCheckpoint(checkpoint *Checkpoint) error {
	result := c.d.Model(&Checkpoint{}).Create(&checkpoint).Commit()
	return result.Error
}

func (c *DbClient) updateCheckpointByType(typ uint16, checkpoint *Checkpoint) error {
	result := c.d.Model(&Checkpoint{}).
		Where("type = ?", typ).
		Updates(map[string]interface{}{
			"block_number": checkpoint.BlockNumber,
		})
	return result.Error
}
