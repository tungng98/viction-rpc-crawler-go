package svc

import (
	"math/big"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
	"github.com/tforce-io/tf-golib/opx"
)

type TraceBlock struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	o   *TraceBlockOptions
	rpc *rpc.EthClient
}

func NewTraceBlock(logger diag.Logger, rpc *rpc.EthClient) *TraceBlock {
	svc := &TraceBlock{
		rpc: rpc,
	}
	svc.i = svc.InitServiceCore("TraceBlock", logger, svc.coreProcessHook)
	svc.o = &TraceBlockOptions{
		MaxRetries: 3,
	}
	return svc
}

func (s *TraceBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "trace_block":
		blockNumber := msg.GetParam("block_number", new(big.Int)).(*big.Int)
		blockTraces, str, err := s.rpc.TraceBlockByNumber(blockNumber)
		retryCount := 0
		for err != nil && retryCount < s.o.MaxRetries {
			blockTraces, str, err = s.rpc.TraceBlockByNumber(blockNumber)
			retryCount++
		}
		result := &TraceBlockResult{
			Number:  blockNumber,
			Data:    blockTraces,
			RawData: str,
			Error:   err,
		}
		s.i.Logger.Infof("%s#%d: Block #%d processed. %s. Retry count = %d.", s.i.ServiceID, workerID, blockNumber.Uint64(),
			opx.Ternary(err == nil, "SUCCESS", "FAILED"),
			retryCount,
		)
		msg.Return(result)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

type TraceBlockOptions struct {
	MaxRetries int
}

type TraceBlockResult struct {
	Number  *big.Int
	Data    rpc.TraceBlockResult
	RawData string
	Error   error
}
