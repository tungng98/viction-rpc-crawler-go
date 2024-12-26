package svc

import (
	"math/big"
	"sync"
	"time"
	"viction-rpc-crawler-go/cache"
	"viction-rpc-crawler-go/rpc"

	"github.com/rs/zerolog"
)

type BlockFetcherCommand struct {
	Command string
	Params  ExecParams
}

type BlockFetcherQueueCommand struct {
	RequestID    string
	BlockNumbers []*big.Int
	Returns      *sync.WaitGroup
}

type GetBlockResult struct {
	BlockNumber *big.Int
	Data        *rpc.Block
	Error       error
}

type BlockFetcherSvc struct {
	i *BlockFetcherSvcInternal
}

type BlockFetcherSvcInternal struct {
	WorkerMainCounter  *WorkerCounter
	WorkerMainCount    uint16
	WorkerQueueCounter *WorkerCounter
	WorkerID           uint64

	MainChan  chan *BlockFetcherCommand
	QueueChan chan *BlockFetcherQueueCommand
	QueueWait map[string]*sync.WaitGroup

	Controller  ServiceController
	Rpc         *rpc.EthClient
	SharedCache cache.MemoryCache
	Logger      *zerolog.Logger
}

func NewBlockFetcherSvc(controller ServiceController, rpc *rpc.EthClient, cache cache.MemoryCache, logger *zerolog.Logger) BlockFetcherSvc {
	svc := BlockFetcherSvc{
		i: &BlockFetcherSvcInternal{
			WorkerMainCounter:  &WorkerCounter{},
			WorkerQueueCounter: &WorkerCounter{},

			MainChan:  make(chan *BlockFetcherCommand, MAIN_CHAN_CAPACITY),
			QueueChan: make(chan *BlockFetcherQueueCommand, SECOND_CHAN_CAPACITY),
			QueueWait: map[string]*sync.WaitGroup{},

			Controller:  controller,
			Rpc:         rpc,
			SharedCache: cache,
			Logger:      logger,
		},
	}
	return svc
}

func (s BlockFetcherSvc) ServiceID() string {
	return "BlockFetcher"
}

func (s BlockFetcherSvc) Controller() ServiceController {
	return s.i.Controller
}

func (s BlockFetcherSvc) SetWorker(workerCount uint16) {
	s.i.WorkerMainCounter.Lock()
	defer s.i.WorkerMainCounter.Unlock()
	if s.i.WorkerMainCounter.ValueNoLock() != s.i.WorkerMainCount {
		return
	}
	if workerCount > s.i.WorkerMainCounter.ValueNoLock() {
		for i := s.i.WorkerMainCounter.ValueNoLock(); i < workerCount; i++ {
			s.i.WorkerID++
			go s.process(s.i.WorkerID)
		}
		s.i.WorkerMainCount = workerCount
	}
	if workerCount < s.i.WorkerMainCounter.ValueNoLock() {
		for i := s.i.WorkerMainCounter.ValueNoLock(); i > workerCount; i-- {
			cmd := &BlockFetcherCommand{
				Command: "exit",
			}
			s.i.MainChan <- cmd
		}
		s.i.WorkerMainCount = workerCount
	}
}

func (s BlockFetcherSvc) WorkerCount() uint16 {
	return s.i.WorkerMainCounter.ValueNoLock() + s.i.WorkerQueueCounter.ValueNoLock()
}

func (s BlockFetcherSvc) Exec(command string, params ExecParams) {
	if command == "exit" {
		return
	}
	if command == "get_blocks" || command == "get_blocks_range" {
		requestID := params["request_id"].(string)
		blockNumbers := []*big.Int{}
		if command == "get_blocks" {
			blockNumbers = params["block_numbers"].([]*big.Int)
		}
		if command == "get_blocks_range" {
			fromBlock := params["from"].(*big.Int).Uint64()
			toBlock := params["to"].(*big.Int).Uint64()
			blockNumbers = make([]*big.Int, toBlock-fromBlock+1)
			index := 0
			for i := fromBlock; i <= toBlock; i++ {
				blockNumbers[index] = new(big.Int).SetUint64(i)
				index++
			}
		}
		s.i.WorkerQueueCounter.Lock()
		{
			defer s.i.WorkerQueueCounter.Unlock()
			if _, ok := s.i.QueueWait[requestID]; ok {
				s.i.Logger.Warn().Msgf("Request existed %s", requestID)
				return
			}
			s.i.QueueWait[requestID] = new(sync.WaitGroup)
			s.i.QueueWait[requestID].Add(len(blockNumbers))
		}
		queueMsg := &BlockFetcherQueueCommand{
			RequestID:    requestID,
			BlockNumbers: blockNumbers,
		}
		if ret, ok := params["returns"]; ok {
			queueMsg.Returns = ret.(*sync.WaitGroup)
		}
		s.i.QueueChan <- queueMsg
		s.i.WorkerID++
		go s.processQueue(s.i.WorkerID)
		return
	}
	msg := &BlockFetcherCommand{
		Command: command,
		Params:  params,
	}
	s.i.MainChan <- msg
}

func (s *BlockFetcherSvc) process(workerID uint64) {
	s.i.WorkerMainCounter.Increase()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockFetcher Process started.")
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.MainChan
		switch msg.Command {
		case "get_block":
			requestID := msg.Params["request_id"].(string)
			index := msg.Params["index"].(int)
			blockNumber := msg.Params["block_number"].(*big.Int)
			block, err := s.i.Rpc.GetBlockByNumber2(blockNumber)
			result := &GetBlockResult{
				BlockNumber: blockNumber,
				Data:        block,
				Error:       err,
			}
			cache.SetArrayItem(s.i.SharedCache, requestID, index, result)
			s.i.Logger.Info().Uint64("WorkerID", workerID).Msgf("Fetched block %d.", blockNumber.Uint64())
			s.i.QueueWait[requestID].Done()
		case "exit":
			status = EXIT_STATE
		}
	}
	s.i.WorkerMainCounter.Decrease()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockFetcher Process exited.")
}

func (s *BlockFetcherSvc) processQueue(workerID uint64) {
	s.i.WorkerQueueCounter.Lock()
	{
		defer s.i.WorkerQueueCounter.Unlock()
		if s.i.WorkerQueueCounter.ValueNoLock() > 0 {
			return
		}
		s.i.WorkerQueueCounter.IncreaseNoLock()
	}

	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockFetcher ProcessQueue started.")
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.QueueChan
		cache.AllocArray(s.i.SharedCache, msg.RequestID, len(msg.BlockNumbers))
		for i := 0; i < len(msg.BlockNumbers); i++ {
			msgi := &BlockFetcherCommand{
				Command: "get_block",
				Params: ExecParams{
					"request_id":   msg.RequestID,
					"index":        i,
					"block_number": msg.BlockNumbers[i],
				},
			}
			s.i.MainChan <- msgi
		}
		startTime := time.Now()
		s.i.QueueWait[msg.RequestID].Wait()
		if msg.Returns != nil {
			msg.Returns.Done()
		}
		s.i.Logger.Info().Msgf("Fetched request %s in %s.", msg.RequestID, time.Since(startTime))
		s.i.WorkerQueueCounter.Lock()
		{
			defer s.i.WorkerQueueCounter.Unlock()
			delete(s.i.QueueWait, msg.RequestID)
			if len(s.i.QueueWait) == 0 {
				status = EXIT_STATE
			}
			s.i.WorkerQueueCounter.DecreaseNoLock()
		}
	}
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockFetcher ProcessQueue exited.")
}
