package svc

import (
	"math/big"
	"strings"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
	"github.com/tforce-io/tf-golib/opx"
)

type TraceBlock struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	o   *NetworkOptions
	rpc *rpc.EthClient
}

func NewTraceBlock(logger diag.Logger, rpc *rpc.EthClient) *TraceBlock {
	svc := &TraceBlock{
		rpc: rpc,
	}
	svc.i = svc.InitServiceCore("TraceBlock", logger, svc.coreProcessHook)
	svc.o = &NetworkOptions{
		MaxRetries:  5,
		MaxRetryGap: 200 * 1000000,
	}
	return svc
}

func (s *TraceBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "trace_block":
		blockNumber := msg.GetParam("block_number", new(big.Int)).(*big.Int)
		blockTraces, str, err := s.rpc.TraceBlockByNumber(blockNumber)
		retryCount := 0
		halfRetry := false
		for err != nil && retryCount < s.o.MaxRetries {
			errStr := err.Error()
			if strings.HasPrefix(err.Error(), "503 Service Unavailable: <html><body><h1>503 Service Unavailable</h1>") {
				if !halfRetry {
					retryCount--
				}
				halfRetry = !halfRetry
			} else {
				s.i.Logger.Warnf("%s#%d: Block #%d retrying. %v", s.i.ServiceID, workerID, blockNumber.Uint64(), errStr)
			}
			s.o.WaitRetryGap()
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

type TraceBlockResult struct {
	Number  *big.Int
	Data    rpc.TraceBlockResult
	RawData string
	Error   error
}
