package db

import (
	"math/big"
	"testing"
)

func TestGetHighestBlock(t *testing.T) {
	db := prepareDatabaseForBlocks()
	defer db.Disconnect()

	tests := []struct {
		Name   string
		Number *big.Int
	}{
		{"highest", big.NewInt(17000)},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			block, err := db.GetHighestIndexBlock()
			if err != nil {
				t.Fatalf("Error while getting highest block number. %v", err)
			}
			if block.BlockNumber != tt.Number.Uint64() {
				t.Fatalf("Highest block number mismatch. Expected '%s' Actual '%d'", tt.Number.String(), block.BlockNumber)
			}
		})
	}
}

func prepareDatabaseForBlocks() *DbClient {
	db, err := Connect(TEST_CONNECTION, "")
	if err != nil {
		panic(err)
	}
	err = db.SaveHighestIndexBlock(big.NewInt(17000))
	if err != nil {
		panic(err)
	}
	return db
}
