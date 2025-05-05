package svc

import (
	"math/big"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type GetBlock struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	o   *GetBlockOptions
	rpc *rpc.EthClient
}

func NewGetBlock(logger diag.Logger, rpc *rpc.EthClient) *GetBlock {
	svc := &GetBlock{
		rpc: rpc,
	}
	svc.i = svc.InitServiceCore("GetBlock", logger, svc.coreProcessHook)
	svc.o = &GetBlockOptions{
		MaxRetries: 3,
	}
	return svc
}

func (s *GetBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_block":
		blockNumber := msg.GetParam("block_number", new(big.Int)).(*big.Int)
		block, str, err := s.rpc.GetBlockByNumber2(blockNumber)
		retryCount := 0
		for err != nil && retryCount < s.o.MaxRetries {
			block, str, err = s.rpc.GetBlockByNumber2(blockNumber)
			retryCount++
		}
		result := &GetBlockResult{
			Number:  blockNumber,
			Data:    block,
			RawData: str,
			Error:   err,
		}
		s.i.Logger.Infof("%s#%d: Block #%d retrieved.", s.i.ServiceID, workerID, blockNumber.Uint64())
		msg.Return(result)
	case "get_block_number":
		head, err := s.rpc.GetBlockNumber()
		retryCount := 0
		for err != nil && retryCount < s.o.MaxRetries {
			head, err = s.rpc.GetBlockNumber()
			retryCount++
		}
		result := &GetBlockNumberResult{
			Number: new(big.Int).SetUint64(head),
			Error:  err,
		}
		msg.Return(result)
	}
	return &multiplex.HookState{Handled: true}
}

type GetBlockOptions struct {
	MaxRetries int
}

type GetBlockResult struct {
	Number  *big.Int
	Data    *rpc.Block
	RawData string
	Error   error
}

type GetBlockNumberResult struct {
	Number *big.Int
	Error  error
}
