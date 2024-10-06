package db

import (
	"math/big"
	"testing"
)

func TestGetTxHash(t *testing.T) {
	db := prepareDatabaseForTxHashes()
	defer db.Disconnect()

	tests := []struct {
		Hash        string
		BlockNumber *big.Int
		BlockHash   string
	}{
		{"0x1b8c30aa5966d89ce4b8a9a32ad7c3a097dbe0f1ec47eaeed0688bfcf1357e0c", big.NewInt(17000), "0xe86731ab1d395ee89b42efa91211f800362e961de4e421f9938b71bc4b508ac1"},
		{"0xbc5c6939407b53375091f2d9276a004d6c5bd64ad603e9d772e4514fd7e8a46e", big.NewInt(17001), "0xfb0cff02600852fd3d105dc6dafb9d2ffdf6a31adc678ace018c1a893513798a"},
	}
	for _, tt := range tests {
		t.Run(tt.Hash, func(t *testing.T) {
			txHash, err := db.GetTxHash(tt.Hash)
			if err != nil {
				t.Fatalf("Error while getting txHash. %v", err)
			}
			if !txHash.BlockNumber.Equals2(tt.BlockNumber) {
				t.Fatalf("Block number mismatch. Expected '%s' Actual '%s'", tt.BlockNumber.String(), txHash.BlockNumber.String())
			}
			if txHash.BlockHash != tt.BlockHash {
				t.Fatalf("Block hash mismatch. Expected '%s' Actual '%s'", tt.BlockHash, txHash.BlockHash)
			}
		})
	}
}

func prepareDatabaseForTxHashes() *DbClient {
	db, err := Connect("mongodb://localhost:27017", "viction_test")
	if err != nil {
		panic(err)
	}
	err = db.SaveTxHash("0x1b8c30aa5966d89ce4b8a9a32ad7c3a097dbe0f1ec47eaeed0688bfcf1357e0c", big.NewInt(17000), "0xe86731ab1d395ee89b42efa91211f800362e961de4e421f9938b71bc4b508ac1")
	if err != nil {
		panic(err)
	}
	err = db.SaveTxHash("0xbc5c6939407b53375091f2d9276a004d6c5bd64ad603e9d772e4514fd7e8a46e", big.NewInt(17001), "0xfb0cff02600852fd3d105dc6dafb9d2ffdf6a31adc678ace018c1a893513798a")
	if err != nil {
		panic(err)
	}
	return db
}
