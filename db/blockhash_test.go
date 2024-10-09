package db

import (
	"math/big"
	"testing"
)

func TestGetBlockHash(t *testing.T) {
	db := prepareDatabaseForBlockHashes()
	defer db.Disconnect()

	tests := []struct {
		Hash        string
		BlockNumber *big.Int
	}{
		{"0xe86731ab1d395ee89b42efa91211f800362e961de4e421f9938b71bc4b508ac1", big.NewInt(17000)},
		{"0xfb0cff02600852fd3d105dc6dafb9d2ffdf6a31adc678ace018c1a893513798a", big.NewInt(17001)},
	}
	for _, tt := range tests {
		t.Run(tt.Hash, func(t *testing.T) {
			txHash, err := db.GetBlockHash(tt.Hash)
			if err != nil {
				t.Fatalf("Error while getting BlockHash. %v", err)
			}
			if !txHash.BlockNumber.Equals2(tt.BlockNumber) {
				t.Fatalf("Block number mismatch. Expected '%s' Actual '%s'", tt.BlockNumber.String(), txHash.BlockNumber.String())
			}
		})
	}
}

func prepareDatabaseForBlockHashes() *DbClient {
	db, err := Connect("mongodb://localhost:27017", "viction_test")
	if err != nil {
		panic(err)
	}
	err = db.SaveBlockHash("0xe86731ab1d395ee89b42efa91211f800362e961de4e421f9938b71bc4b508ac1", big.NewInt(17000))
	if err != nil {
		panic(err)
	}
	err = db.SaveBlockHash("0xfb0cff02600852fd3d105dc6dafb9d2ffdf6a31adc678ace018c1a893513798a", big.NewInt(17001))
	if err != nil {
		panic(err)
	}
	return db
}
