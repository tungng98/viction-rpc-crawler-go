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

func NewGetBlock(logger diag.Logger) *GetBlock {
	svc := &GetBlock{}
	svc.i = svc.InitServiceCore("GetBlock", logger, svc.coreProcessHook)
	svc.o = &GetBlockOptions{
		MaxRetries: 3,
	}
	return svc
}

func (s *GetBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_block":
		blockNumber := msg.Params["block_number"].(*big.Int)
		block, err := s.rpc.GetBlockByNumber2(blockNumber)
		retryCount := 0
		for err != nil && retryCount < s.o.MaxRetries {
			block, err = s.rpc.GetBlockByNumber2(blockNumber)
			retryCount++
		}
		result := &GetBlockResult{
			Number: blockNumber,
			Data:   block,
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
	Number *big.Int
	Data   *rpc.Block
	Error  error
}
