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
			ethutil.HexToBytes(ethutil.RandomTxHash()),
			new(big.Int).SetUint64(currentBlockNum),
			ethutil.HexToBytes(ethutil.RandomBlockHash()),
			new(big.Int).SetUint64(prevBlockNum),
			ethutil.HexToBytes(ethutil.RandomBlockHash()),
		)
		if err != nil {
			t.Fatalf("Error while getting saving issue. %v", err)
		}
		_, err = db.SaveIssues([]*Issue{
			NewDuplicatedTxHashIssue(
				ethutil.HexToBytes(ethutil.RandomTxHash()),
				new(big.Int).SetUint64(ethutil.RandomNumber(0, ^uint64(0))),
				ethutil.HexToBytes(ethutil.RandomBlockHash()),
				new(big.Int).SetUint64(ethutil.RandomNumber(0, ^uint64(0))),
				ethutil.HexToBytes(ethutil.RandomBlockHash()),
			),
			NewDuplicatedBlockHashIssue(
				ethutil.HexToBytes(ethutil.RandomTxHash()),
				new(big.Int).SetUint64(ethutil.RandomNumber(0, ^uint64(0))),
				new(big.Int).SetUint64(ethutil.RandomNumber(0, ^uint64(0))),
			),
		})
		if err != nil {
			t.Fatalf("Error while getting saving issue. %v", err)
		}
	})
}

func prepareDatabaseForIssues() *DbClient {
	db, err := Connect(TEST_CONNECTION, "")
	if err != nil {
		panic(err)
	}
	return db
}
