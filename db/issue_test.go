package db

import (
	"math/big"
	"testing"
	"viction-rpc-crawler-go/x/ethutil"
)

func TestInsertIssue(t *testing.T) {
	db := prepareDatabaseForIssues()
	defer db.Disconnect()

	t.Run("duplicated_hash", func(t *testing.T) {
		currentBlockNum := ethutil.RandomNumber(0, ^uint64(0))
		prevBlockNum := ethutil.RandomNumber(0, currentBlockNum)
		err := db.SaveDuplicatedTxHashIssue(
			ethutil.RandomTxHash(),
			new(big.Int).SetUint64(currentBlockNum),
			ethutil.RandomBlockHash(),
			new(big.Int).SetUint64(prevBlockNum),
			ethutil.RandomBlockHash(),
		)
		if err != nil {
			t.Fatalf("Error while getting saving issue. %v", err)
		}
	})
}

func prepareDatabaseForIssues() *DbClient {
	db, err := Connect("mongodb://localhost:27017", "viction_test")
	if err != nil {
		panic(err)
	}
	return db
}
