package db

type Block struct {
	ID                    uint64 `gorm:"column:id;primaryKey"`
	Hash                  []byte `gorm:"column:hash;size:32;uniqueIndex"`
	ParentHash            []byte `gorm:"column:parent_hash;size:32"`
	StateRoot             []byte `gorm:"column:state_root;size:32"`
	TransactionsRoot      []byte `gorm:"column:transaction_root;size:32"`
	ReceiptsRoot          []byte `gorm:"column:receipts_root;size:32"`
	Timestamp             int64  `gorm:"column:timestamp"`
	Size                  uint16 `gorm:"column:size"`
	GasLimit              uint64 `gorm:"column:gas_limit"`
	GasUsed               uint64 `gorm:"column:gas_used"`
	TotalDifficulty       uint64 `gorm:"column:total_difficult"`
	TransactionCount      uint16 `gorm:"column:transaction_count"`
	TransactionCountDebug uint16 `gorm:"column:transaction_count_debug"`
	Transaction           []Transaction
}
