package db

import (
	"math/big"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID               uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	Hash             string          `gorm:"column:hash;length:32;uniqueIndex"`
	BlockID          uint64          `gorm:"column:block_id;index"`
	BlockHash        string          `gorm:"column:block_hash;length:32;index"`
	TransactionIndex uint16          `gorm:"column:transaction_index"`
	From             string          `gorm:"column:from;length:20"`
	To               string          `gorm:"column:to;length:20"`
	Value            decimal.Decimal `gorm:"column:value;type:decimal(78,0)"`
	Nonce            uint64          `gorm:"column:nonce"`
	Gas              uint64          `gorm:"column:gas"`
	GasPrice         decimal.Decimal `gorm:"column:gas_price;type:decimal(78,0)"`
}

func NewTransaction(hash string, blockNumber *big.Int, blockHash string, transactionIndex uint16, from, to string, value *big.Int, nonce uint64, gas uint64, gasPrice *big.Int) *Transaction {
	return &Transaction{
		Hash:             hash,
		BlockID:          blockNumber.Uint64(),
		BlockHash:        blockHash,
		TransactionIndex: transactionIndex,
		From:             from,
		To:               to,
		Value:            decimal.NewFromBigInt(value, 0),
		Nonce:            nonce,
		Gas:              gas,
		GasPrice:         decimal.NewFromBigInt(gasPrice, 0),
	}
}

func (c *DbClient) GetTransaction(hash string) (*Transaction, error) {
	return c.findTransactonByHash(hash)
}

func (c *DbClient) GetTransactions(hashes []string) ([]*Transaction, error) {
	return c.findTransactonsByHashes(hashes)
}

func (c *DbClient) SaveTransaction(hash string, newTransaction *Transaction) error {
	transaction, err := c.GetTransaction(hash)
	if err != nil {
		return err
	}
	if transaction == nil {
		return c.insertTxHash(transaction)
	}
	if transaction.BlockID == newTransaction.BlockID {
		return nil
	}
	err = c.SaveDuplicatedTxHashIssue(hash, newTransaction.BlockID, newTransaction.BlockHash, transaction.BlockID, transaction.BlockHash)
	if err != nil {
		return err
	}
	return c.updateTransactionByHash(hash, newTransaction)
}

func (c *DbClient) SaveTransactions(newTransactions []*Transaction, changedTransactions []*Transaction) error {
	return c.writeTransactions(newTransactions, changedTransactions)
}

func (c *DbClient) findTransactonByHash(hash string) (*Transaction, error) {
	var doc *Transaction
	result := c.d.Model(&Transaction{}).
		Where("hash = ?", hash).
		First(&doc)
	if c.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (c *DbClient) findTransactonsByHashes(hashes []string) ([]*Transaction, error) {
	var docs []*Transaction
	result := c.d.Model(&Transaction{}).
		Where("hash IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

func (c *DbClient) insertTxHash(transaction *Transaction) error {
	result := c.d.Create(transaction)
	return result.Error
}

func (c *DbClient) updateTransactionByHash(hash string, transaction *Transaction) error {
	result := c.d.Model(&Transaction{}).
		Where("hash = ?", hash).
		Updates(map[string]interface{}{
			"block_id":          transaction.BlockID,
			"block_hash":        transaction.BlockHash,
			"transaction_index": transaction.TransactionIndex,
			"from":              transaction.From,
			"to":                transaction.To,
			"value":             transaction.Value,
			"nonce":             transaction.Nonce,
			"gas":               transaction.Gas,
			"gas_price":         transaction.GasPrice,
		})
	return result.Error
}

func (c *DbClient) writeTransactions(newTransactions []*Transaction, changedTransactions []*Transaction) error {
	tx := c.d.Begin()
	for _, transaction := range newTransactions {
		result := tx.Create(transaction)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	for _, transaction := range changedTransactions {
		result := tx.Model(&Transaction{}).
			Where("hash = ?", transaction.Hash).
			Updates(map[string]interface{}{
				"block_id":          transaction.BlockID,
				"block_hash":        transaction.BlockHash,
				"transaction_index": transaction.TransactionIndex,
				"from":              transaction.From,
				"to":                transaction.To,
				"value":             transaction.Value,
				"nonce":             transaction.Nonce,
				"gas":               transaction.Gas,
				"gas_price":         transaction.GasPrice,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}
