package svc

import (
	"math/big"

	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"
)

type IndexBlockTxService struct {
	DbConnStr string
	DbName    string
	RpcUrl    string
}

func (s *IndexBlockTxService) Exec() {
	db, err := db.Connect(s.DbConnStr, s.DbName)
	if err != nil {
		panic(err)
	}
	defer db.Disconnect()
	rpc, err := rpc.Connect(s.RpcUrl)
	if err != nil {
		panic(err)
	}

	checkpoint, err := db.GetHighestBlock()
	if err != nil {
		panic(err)
	}
	blockNum := big.NewInt(1)
	if checkpoint != nil {
		blockNum = checkpoint.BlockNumber.N
	}
	for {
		finality, err := rpc.GetBlockFinalityByNumber(blockNum)
		if err != nil {
			panic(err)
		}
		if finality < 75 {
			break
		}
		block, err := rpc.GetBlockByNumber(blockNum)
		if err != nil {
			panic(err)
		}
		db.SaveBlockHash(block.Hash().Hex(), blockNum)
		for _, tx := range block.Transactions() {
			db.SaveTxHash(tx.Hash().Hex(), blockNum, block.Hash().Hex())
		}
		db.SaveHighestBlock(blockNum)
		blockNum = new(big.Int).Add(blockNum, big.NewInt(1))
	}
}
