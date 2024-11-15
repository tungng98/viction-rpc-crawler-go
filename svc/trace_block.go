package svc

import (
	"math/big"
	"strings"
	"sync"
	"time"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"

	"github.com/rs/zerolog"
)

type TraceBlockService struct {
	DbConnStr          string
	DbName             string
	RpcUrl             string
	StartBlock         int64
	EndBlock           int64
	UseCheckpointBlock bool
	SaveDebugData      bool
	BatchSize          int
	WorkerCount        int
	Logger             *zerolog.Logger

	workers *TraceBlockDataQueue
}

func (s *TraceBlockService) Exec() {
	s.init()
	db, err := db.Connect(s.DbConnStr, s.DbName)
	if err != nil {
		panic(err)
	}
	defer db.Disconnect()
	rpc, err := rpc.Connect(s.RpcUrl)
	if err != nil {
		panic(err)
	}
	if s.workers == nil {
		s.workers = &TraceBlockDataQueue{
			chanBlockNumbers: make(chan *TraceBlockDataItem, s.WorkerCount),
			client:           rpc,
			logger:           s.Logger,
		}
		for i := 1; i <= s.WorkerCount; i++ {
			go s.workers.Start()
		}
	}
	startBlock := big.NewInt(s.StartBlock)
	if s.UseCheckpointBlock {
		highestBlock, err := db.GetHighestTraceBlock()
		if err != nil {
			panic(err)
		}
		if highestBlock != nil {
			startBlock = highestBlock.BlockNumber.N
		}
	}
	endBlock := big.NewInt(s.EndBlock)
	for startBlock.Cmp(endBlock) < 0 {
		blockTraces, _ := s.getBlockData(startBlock)
		startTime := time.Now()
		oldStartBlockNumber := startBlock
		endBlockNumber := new(big.Int).Add(oldStartBlockNumber, big.NewInt(int64(s.BatchSize)-1))
		startBlock = new(big.Int).Add(startBlock, big.NewInt(int64(len(blockTraces))))
		if s.UseCheckpointBlock {
			err = db.SaveHighestTraceBlock(startBlock)
			if err != nil {
				panic(err)
			}
		}
		s.Logger.Info().
			Msgf("Persisted Batch #%d-%d in %v", oldStartBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
	}
}

func (s *TraceBlockService) init() {
	if s.WorkerCount == 0 {
		s.WorkerCount = 1
	}
	if s.BatchSize == 0 {
		s.BatchSize = 1
	}
}

func (s *TraceBlockService) getBlockData(startBlockNumber *big.Int) ([]*TraceBlockDataResult, error) {
	s.workers.BlockData = make([]*TraceBlockDataResult, s.BatchSize)
	s.workers.ChanCompleteSignal = new(sync.WaitGroup)
	s.workers.ChanCompleteSignal.Add(s.BatchSize)
	startTime := time.Now()
	for i := 0; i < s.BatchSize; i++ {
		s.workers.Equneue(big.NewInt(startBlockNumber.Int64()+int64(i)), i)
	}
	s.workers.ChanCompleteSignal.Wait()
	endBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(int64(s.BatchSize)-1))
	s.Logger.Info().Msgf("Traced Batch #%d-%d in %v", startBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
	return s.workers.BlockData, s.workers.Error
}

type TraceBlockBatchDataResult struct {
	Traces []*TraceBlockDataResult
	Issues []*db.Issue
}

type TraceBlockDataQueue struct {
	chanBlockNumbers   chan *TraceBlockDataItem
	client             *rpc.EthClient
	logger             *zerolog.Logger
	BlockData          []*TraceBlockDataResult
	ChanCompleteSignal *sync.WaitGroup
	Error              error
}

type TraceBlockDataItem struct {
	blockNumber *big.Int
	index       int
	retry       int
}

type TraceBlockDataResult struct {
	BlockNumber *big.Int
	BlockHash   string
	Traces      []*rpc.TraceTransactionResult
}

func (q *TraceBlockDataQueue) Start() {
	for {
		select {
		case data := <-q.chanBlockNumbers:
			startTime := time.Now()
			var err error
			traces, err := q.client.TraceBlockByNumber(data.blockNumber)
			if err != nil {
				q.BlockData[data.index] = nil
				q.Error = err
				q.logger.Err(err).Msgf("Traced block #%d in %v: ERR", data.blockNumber.Int64(), time.Since(startTime))
				if data.retry < 3 {
					if !is502Error(err) {
						data.retry++
					}
					q.chanBlockNumbers <- data // re-enqueue to fix errorq.chanBlockNumbers <- data // re-enqueue to fix error
				} else {
					q.ChanCompleteSignal.Done()
				}
			} else {
				q.BlockData[data.index] = &TraceBlockDataResult{
					BlockNumber: data.blockNumber,
					Traces:      traces,
				}
				q.logger.Debug().Msgf("Traced block #%d in %v", data.blockNumber.Int64(), time.Since(startTime))
				q.ChanCompleteSignal.Done()
			}
		}
	}
}

func (q *TraceBlockDataQueue) Equneue(blockNumber *big.Int, index int) {
	q.chanBlockNumbers <- &TraceBlockDataItem{blockNumber, index, 0}
}

func is502Error(err error) bool {
	errMsg := err.Error()
	if strings.Contains(errMsg, "502 Bad Gateway") {
		return true
	}
	if strings.Contains(errMsg, ": EOF") {
		return true
	}
	return false
}
