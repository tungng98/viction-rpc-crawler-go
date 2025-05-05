package db

import (
	"testing"

	"github.com/tforce-io/tf-golib/random/pseudorng"
)

func TestInsertIssue(t *testing.T) {
	db := prepareDatabaseForIssues()
	defer db.Disconnect()

	t.Run("duplicated_hash", func(t *testing.T) {
		currentBlockNum := pseudorng.Uint64r(0, ^uint64(0))
		prevBlockNum := pseudorng.Uint64r(0, currentBlockNum)
		err := db.SaveDuplicatedTxHashIssue(
			pseudorng.Hex(32),
			currentBlockNum,
			pseudorng.Hex(32),
			prevBlockNum,
			pseudorng.Hex(32),
		)
		if err != nil {
			t.Fatalf("Error while getting saving issue. %v", err)
		}
		err = db.SaveIssues([]*Issue{
			NewDuplicatedTxHashIssue(
				pseudorng.Hex(32),
				pseudorng.Uint64r(0, ^uint64(0)),
				pseudorng.Hex(32),
				pseudorng.Uint64r(0, ^uint64(0)),
				pseudorng.Hex(32),
			),
			NewReorgBlockIssue(
				pseudorng.Uint64r(0, ^uint64(0)),
				pseudorng.Hex(32),
				pseudorng.Hex(32),
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
