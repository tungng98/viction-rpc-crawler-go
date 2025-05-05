package svc

import (
	"math/big"
	"sync"
	"time"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type GetBlocks struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

func NewGetBlocks(logger diag.Logger) *GetBlocks {
	svc := &GetBlocks{}
	svc.i = svc.InitServiceCore("GetBlocks", logger, svc.coreProcessHook)
	return svc
}

func (s *GetBlocks) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_blocks":
	case "get_blocks_range":
		startTime := time.Now()
		requests := []multiplex.ExecParams{}
		signal := new(sync.WaitGroup)
		if msg.Command == "get_blocks" {
			blockNumbers := msg.GetParam("block_numbers", []*big.Int{}).([]*big.Int)
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
			fromBlockNumber := msg.GetParam("from_block_number", new(big.Int)).(*big.Int)
			toBlockNumber := msg.GetParam("to_block_number", new(big.Int)).(*big.Int)
			for blockNumber := fromBlockNumber; blockNumber.Cmp(toBlockNumber) <= 0; blockNumber.Set(new(big.Int).Add(blockNumber, big.NewInt(1))) {
				request := multiplex.ExecParams{
					"block_number": new(big.Int).Set(blockNumber),
				}
				request.ExpectReturnCustomSignal(signal)
				requests = append(requests, request)
			}
		}
		results := &GetBlocksResult{
			Data: make([]*GetBlockResult, len(requests)),
		}
		if len(requests) > 0 {
			signal.Add(len(requests))
			for _, request := range requests {
				s.Dispatch("GetBlock", "get_block", request)
			}
			signal.Wait()
			for i, request := range requests {
				results.Data[i] = request.ReturnResult().(*GetBlockResult)
			}
		}
		s.i.Logger.Infof("%s#%d: %d blocks retrieved in %v.", s.i.ServiceID, workerID, len(requests), time.Since(startTime))
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
