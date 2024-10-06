package db

import (
	"math/big"
	"testing"
)

func TestGetHighestBlock(t *testing.T) {
	db := prepareDatabaseForBlocks()

	tests := []struct {
		Name   string
		Number *big.Int
	}{
		{"highest", big.NewInt(17000)},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			block, err := db.GetHighestBlock()
			if err != nil {
				t.Fatalf("Error while getting highest block number. %v", err)
			}
			if block.Number.N.Cmp(tt.Number) != 0 {
				t.Fatalf("Highest block number mismatch. Expected '%s' Actual '%s'", tt.Number.String(), block.Number.String())
			}
		})
	}
}

func prepareDatabaseForBlocks() *DbClient {
	db, err := Connect("mongodb://localhost:27017", "viction_test")
	if err != nil {
		panic(err)
	}
	err = db.SaveHighestBlock(big.NewInt(17000))
	if err != nil {
		panic(err)
	}
	return db
}
