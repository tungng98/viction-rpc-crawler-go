package svc

import (
	"math/big"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type DownloadBlock struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

func NewDownloadBlock(logger diag.Logger) *DownloadBlock {
	svc := &DownloadBlock{}
	svc.i = svc.InitServiceCore("DownloadBlock", logger, svc.coreProcessHook)
	return svc
}

func (s *DownloadBlock) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "download_blocks":
		fromBlockNumber := msg.GetParam("from_block_number", new(big.Int)).(*big.Int)
		toBlockNumber := msg.GetParam("to_block_number", new(big.Int)).(*big.Int)
		batchSize := msg.GetParam("batch_size", 1).(int)
		root := msg.GetParam("root", "").(string)
		s.downloadBlocks(workerID, fromBlockNumber, toBlockNumber, batchSize, root)
		msg.Return(true)
	case "download_block_traces":
		fromBlockNumber := msg.GetParam("from_block_number", new(big.Int)).(*big.Int)
		toBlockNumber := msg.GetParam("to_block_number", new(big.Int)).(*big.Int)
		batchSize := msg.GetParam("batch_size", 1).(int)
		root := msg.GetParam("root", "").(string)
		s.downloadBlockTraces(workerID, fromBlockNumber, toBlockNumber, batchSize, root)
		msg.Return(true)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *DownloadBlock) downloadBlocks(workerID uint64, from, to *big.Int, batch int, root string) {
	s.i.Logger.Infof("%s#%d: Block download started.", s.ServiceID(), workerID)
	batchStartBlockNumber := new(big.Int).Set(from)
	finalBlockNumber := new(big.Int).Set(to)
	for batchStartBlockNumber.Cmp(finalBlockNumber) < 0 {
		batchEndBlockNumber := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(batch)-1))
		if batchEndBlockNumber.Cmp(finalBlockNumber) > 0 {
			batchEndBlockNumber.Set(finalBlockNumber)
		}
		getBlocksRequest := multiplex.ExecParams{
			"from_block_number": batchStartBlockNumber,
			"to_block_number":   batchEndBlockNumber,
		}
		getBlocksRequest.ExpectReturn()
		s.Dispatch("GetBlocks", "get_blocks_range", getBlocksRequest)
		getBlocksResponse := getBlocksRequest.WaitForReturn().(*GetBlocksResult)
		getBlockResults := []*GetBlockResult{}
		for _, blockResult := range getBlocksResponse.Data {
			if blockResult.Error != nil {
				continue
			}
			getBlockResults = append(getBlockResults, blockResult)
		}
		writeBlockRequest := multiplex.ExecParams{
			"blocks": getBlockResults,
			"root":   root,
		}
		writeBlockRequest.ExpectReturn()
		s.Dispatch("WriteFileSystem", "eth_getBlockByNumber", writeBlockRequest)
		writeBlockRequest.Wait()
		batchStartBlockNumber = new(big.Int).Add(batchEndBlockNumber, big.NewInt(1))
	}
}

func (s *DownloadBlock) downloadBlockTraces(workerID uint64, from, to *big.Int, batch int, root string) {
	s.i.Logger.Infof("%s#%d: Block traces download started.", s.ServiceID(), workerID)
	batchStartBlockNumber := new(big.Int).Set(from)
	finalBlockNumber := new(big.Int).Set(to)
	for batchStartBlockNumber.Cmp(finalBlockNumber) < 0 {
		batchEndBlockNumber := new(big.Int).Add(batchStartBlockNumber, big.NewInt(int64(batch)-1))
		if batchEndBlockNumber.Cmp(finalBlockNumber) > 0 {
			batchEndBlockNumber.Set(finalBlockNumber)
		}
		traceBlocksRequest := multiplex.ExecParams{
			"from_block_number": batchStartBlockNumber,
			"to_block_number":   batchEndBlockNumber,
		}
		traceBlocksRequest.ExpectReturn()
		s.Dispatch("TraceBlocks", "trace_blocks_range", traceBlocksRequest)
		traceBlocksResponse := traceBlocksRequest.WaitForReturn().(*TraceBlocksResult)
		traceBlockResults := []*TraceBlockResult{}
		for _, traceBlockResult := range traceBlocksResponse.Data {
			if traceBlockResult.Error != nil {
				continue
			}
			traceBlockResults = append(traceBlockResults, traceBlockResult)
		}
		writeBlockTracesRequest := multiplex.ExecParams{
			"block_traces": traceBlockResults,
			"root":         root,
		}
		writeBlockTracesRequest.ExpectReturn()
		s.Dispatch("WriteFileSystem", "debug_traceBlockByNumber", writeBlockTracesRequest)
		writeBlockTracesRequest.Wait()
		batchStartBlockNumber = new(big.Int).Add(batchEndBlockNumber, big.NewInt(1))
	}
}
