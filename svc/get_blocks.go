package svc

import (
	"math/big"
	"sync"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type GetBlocks struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
	o *GetBlocksOptions
}

func NewGetBlocks(logger diag.Logger) *GetBlocks {
	svc := &GetBlocks{}
	svc.i = svc.InitServiceCore("GetBlocks", logger, svc.coreProcessHook)
	svc.o = &GetBlocksOptions{
		MaxRetries: 3,
	}
	return svc
}

func (s *GetBlocks) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_blocks":
	case "get_blocks_range":
		requests := []multiplex.ExecParams{}
		signal := new(sync.WaitGroup)
		if msg.Command == "get_blocks" {
			blockNumbers := msg.Params["block_numbers"].([]*big.Int)
			for _, blockNumber := range blockNumbers {
				request := multiplex.ExecParams{
					"block_number": blockNumber,
					"signal":       signal,
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		if msg.Command == "get_blocks_range" {
			fromBlockNumber := msg.Params["from_block_number"].(*big.Int)
			toBlockNumber := msg.Params["to_block_number"].(*big.Int)
			for blockNumber := fromBlockNumber; blockNumber.Cmp(toBlockNumber) <= 0; blockNumber.Set(new(big.Int).Add(blockNumber, big.NewInt(1))) {
				request := multiplex.ExecParams{
					"block_number": blockNumber,
					"signal":       signal,
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		signal.Add(len(requests))
		for _, request := range requests {
			s.Dispatch("GetBlock", "get_block", request)
		}
		signal.Wait()
		results := &GetBlocksResult{
			Data: make([]*GetBlockResult, len(requests)),
		}
		for i, request := range requests {
			result := request["result"].(*GetBlockResult)
			results.Data[i] = result
		}
		msg.Return(results)
	}
	return &multiplex.HookState{Handled: true}
}

type GetBlocksOptions struct {
	MaxRetries int
}

type GetBlocksResult struct {
	Data []*GetBlockResult
}
