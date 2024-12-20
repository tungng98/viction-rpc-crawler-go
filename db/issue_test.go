package db

import (
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
			currentBlockNum,
			ethutil.RandomBlockHash(),
			prevBlockNum,
			ethutil.RandomBlockHash(),
		)
		if err != nil {
			t.Fatalf("Error while getting saving issue. %v", err)
		}
		err = db.SaveIssues([]*Issue{
			NewDuplicatedTxHashIssue(
				ethutil.RandomTxHash(),
				ethutil.RandomNumber(0, ^uint64(0)),
				ethutil.RandomBlockHash(),
				ethutil.RandomNumber(0, ^uint64(0)),
				ethutil.RandomBlockHash(),
			),
			NewReorgBlockIssue(
				ethutil.RandomNumber(0, ^uint64(0)),
				ethutil.RandomTxHash(),
				ethutil.RandomTxHash(),
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
