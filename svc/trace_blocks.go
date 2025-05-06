package svc

import (
	"math/big"
	"sync"
	"time"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type TraceBlocks struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

func NewTraceBlocks(logger diag.Logger) *TraceBlocks {
	svc := &TraceBlocks{}
	svc.i = svc.InitServiceCore("TraceBlocks", logger, svc.coreProcessHook)
	return svc
}

func (s *TraceBlocks) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "trace_blocks":
	case "trace_blocks_range":
		s.i.Logger.Infof("%s#%d: %s started.", s.i.ServiceID, workerID, msg.Command)
		startTime := time.Now()
		requests := []multiplex.ExecParams{}
		signal := new(sync.WaitGroup)
		if msg.Command == "trace_blocks" {
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
		if msg.Command == "trace_blocks_range" {
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
		results := &TraceBlocksResult{
			Data: make([]*TraceBlockResult, len(requests)),
		}
		if len(requests) > 0 {
			signal.Add(len(requests))
			for _, request := range requests {
				s.Dispatch("TraceBlock", "trace_block", request)
			}
			signal.Wait()
			for i, request := range requests {
				results.Data[i] = request.ReturnResult().(*TraceBlockResult)
			}
		}
		s.i.Logger.Infof("%s#%d: %d blocks traced in %v.", s.i.ServiceID, workerID, len(requests), time.Since(startTime))
		msg.Return(results)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

type TraceBlocksResult struct {
	Data []*TraceBlockResult
}
