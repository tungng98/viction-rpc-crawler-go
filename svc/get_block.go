package svc

import (
	"math/big"
	"strings"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
	"github.com/tforce-io/tf-golib/opx"
)

type GetBlock struct {
	multiplex.ServiceCore
	i   *multiplex.ServiceCoreInternal
	o   *NetworkOptions
	rpc *rpc.EthClient
}

func NewGetBlock(logger diag.Logger, rpc *rpc.EthClient) *GetBlock {
	svc := &GetBlock{
		rpc: rpc,
	}
	svc.i = svc.InitServiceCore("GetBlock", logger, svc.coreProcessHook)
	svc.o = &NetworkOptions{
		MaxRetries:  3,
		MaxRetryGap: 200 * 1000000,
	}
	return svc
}

func (s *GetBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "get_block":
		blockNumber := msg.GetParam("block_number", new(big.Int)).(*big.Int)
		block, str, err := s.rpc.GetBlockByNumber2(blockNumber)
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
				s.i.Logger.Warnf("%s#%02d: Block #%d retrying. %v", s.i.ServiceID, workerID, blockNumber.Uint64(), errStr)
			}
			s.o.WaitRetryGap()
			block, str, err = s.rpc.GetBlockByNumber2(blockNumber)
			retryCount++
		}
		result := &GetBlockResult{
			Number:  blockNumber,
			Data:    block,
			RawData: str,
			Error:   err,
		}
		s.i.Logger.Infof("%s#%02d: Block #%d processed. %s. Retry count = %d.", s.i.ServiceID, workerID, blockNumber.Uint64(),
			opx.Ternary(err == nil, "SUCCESS", "FAILED"),
			retryCount,
		)
		msg.Return(result)
	case "get_block_number":
		head, err := s.rpc.GetBlockNumber()
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
				s.i.Logger.Warnf("%s#%02d: Retrying. %v", s.i.ServiceID, workerID, errStr)
			}
			s.o.WaitRetryGap()
			head, err = s.rpc.GetBlockNumber()
			retryCount++
		}
		result := &GetBlockNumberResult{
			Number: new(big.Int).SetUint64(head),
			Error:  err,
		}
		msg.Return(result)
	default:
		s.i.Logger.Warnf("%s#%02d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
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
