package db

type Transaction struct {
	ID               uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Hash             []byte `gorm:"column:hash;size:32;uniqueIndex"`
	BlockID          uint64 `gorm:"column:block_id;index"`
	TransactionIndex uint16 `gorm:"column:transaction_index"`
	From             []byte `gorm:"column:from;size:20;unique"`
	To               []byte `gorm:"column:to;size:20;unique"`
	Value            uint64 `gorm:"column:value"`
	Nonce            uint64 `gorm:"column:nonce"`
	Gas              uint64 `gorm:"column:gas"`
	GasPrice         uint64 `gorm:"column:gas_price"`
	Block            Block
}
