package svc

import (
	"math/big"
	"sync"
	"time"

	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
)

type IndexBlockTxService struct {
	DbConnStr       string
	RpcUrl          string
	StartBlock      int64
	EndBlock        int64
	UseHighestBlock bool
	BatchSize       int
	WorkerCount     int
	Logger          *zerolog.Logger

	workers *GetBlockDataQueue
}

func (s *IndexBlockTxService) Exec() {
	s.init()
	db, err := db.Connect(s.DbConnStr, "")
	if err != nil {
		panic(err)
	}
	defer db.Disconnect()
	rpc, err := rpc.Connect(s.RpcUrl)
	if err != nil {
		panic(err)
	}
	if s.workers == nil {
		s.workers = &GetBlockDataQueue{
			chanBlockNumbers: make(chan *GetBlockDataItem, s.WorkerCount),
			client:           rpc,
			logger:           s.Logger,
		}
		for i := 1; i <= s.WorkerCount; i++ {
			go s.workers.Start()
		}
	}
	startBlock := big.NewInt(s.StartBlock)
	if s.UseHighestBlock {
		highestBlock, err := db.GetHighestIndexBlock()
		if err != nil {
			panic(err)
		}
		if highestBlock != nil {
			startBlock = new(big.Int).SetUint64(highestBlock.BlockNumber)
		}
	}
	endBlock := big.NewInt(s.EndBlock)
	for startBlock.Cmp(endBlock) <= 0 {
		blocks, err := s.getBlockData(startBlock)
		if err != nil {
			panic(err)
		}
		startTime := time.Now()
		batchData, err := s.prepareBatchData(db, blocks)
		if err != nil {
			panic(err)
		}
		if len(batchData.NewBlocks)+len(batchData.ChangedBlocks) > 0 {
			_, err = db.SaveBlockHashes(batchData.NewBlocks, batchData.ChangedBlocks)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.NewTxs)+len(batchData.ChangedTxs) > 0 {
			_, err = db.SaveTxHashes(batchData.NewTxs, batchData.ChangedTxs)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.Issues) > 0 {
			_, err = db.SaveIssues(batchData.Issues)
			if err != nil {
				panic(err)
			}
		}
		oldStartBlockNumber := startBlock
		endBlockNumber := new(big.Int).Add(oldStartBlockNumber, big.NewInt(int64(s.BatchSize)-1))
		startBlock = new(big.Int).Add(startBlock, big.NewInt(int64(len(blocks))))
		err = db.SaveHighestIndexBlock(startBlock)
		if err != nil {
			panic(err)
		}
		s.Logger.Info().
			Int("NewBlocksCount", len(batchData.NewBlocks)).
			Int("ChangedBlocksCount", len(batchData.ChangedBlocks)).
			Int("NewTxsCount", len(batchData.NewTxs)).
			Int("ChangedTxsCount", len(batchData.ChangedTxs)).
			Int("IssuesCount", len(batchData.Issues)).
			Msgf("Persisted Batch #%d-%d in %v", oldStartBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
	}
}

func (s *IndexBlockTxService) init() {
	if s.WorkerCount == 0 {
		s.WorkerCount = 1
	}
	if s.BatchSize == 0 {
		s.BatchSize = 1
	}
}

func (s *IndexBlockTxService) getBlockData(startBlockNumber *big.Int) ([]*types.Block, error) {
	s.workers.BlockData = make([]*types.Block, s.BatchSize)
	s.workers.ChanCompleteSignal = new(sync.WaitGroup)
	s.workers.ChanCompleteSignal.Add(s.BatchSize)
	startTime := time.Now()
	for i := 0; i < s.BatchSize; i++ {
		s.workers.Equneue(big.NewInt(startBlockNumber.Int64()+int64(i)), i)
	}
	s.workers.ChanCompleteSignal.Wait()
	endBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(int64(s.BatchSize)-1))
	s.Logger.Info().Msgf("Fetched Batch #%d-%d in %v", startBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
	return s.workers.BlockData, s.workers.Error
}

type IndexBlockBatchDataResult struct {
	NewBlocks     []*db.BlockHash
	ChangedBlocks []*db.BlockHash
	NewTxs        []*db.TxHash
	ChangedTxs    []*db.TxHash
	Issues        []*db.Issue
}

func (s *IndexBlockTxService) prepareBatchData(dbc *db.DbClient, blocks []*types.Block) (*IndexBlockBatchDataResult, error) {
	result := &IndexBlockBatchDataResult{
		NewBlocks:     []*db.BlockHash{},
		ChangedBlocks: []*db.BlockHash{},
		NewTxs:        []*db.TxHash{},
		ChangedTxs:    []*db.TxHash{},
		Issues:        []*db.Issue{},
	}

	blockHashes := []string{}
	txHashes := []string{}
	issues := []*db.Issue{}
	for _, block := range blocks {
		blockHashes = append(blockHashes, block.Hash().Hex())
		for _, tx := range block.Transactions() {
			txHashes = append(txHashes, tx.Hash().Hex())
		}
	}

	changedBlocks, err := dbc.GetBlockHashes(blockHashes)
	if err != nil {
		return nil, err
	}
	changedTxs, err := dbc.GetTxHashes(txHashes)
	if err != nil {
		return nil, err
	}
	newBlockMap := make(map[string]*db.BlockHash)
	changedBlockMap := make(map[string]*db.BlockHash)
	newTxMap := make(map[string]*db.TxHash)
	changedTxMap := make(map[string]*db.TxHash)
	for _, block := range changedBlocks {
		changedBlockMap[block.Hash] = block
	}
	for _, tx := range changedTxs {
		changedTxMap[tx.Hash] = tx
	}
	for _, block := range blocks {
		blockHash := block.Hash().Hex()
		blockNumber := block.Number()
		if cblock, ok := changedBlockMap[blockHash]; ok {
			if !cblock.BlockNumber.Equals2(blockNumber) {
				issue := db.NewDuplicatedBlockHashIssue(blockHash, blockNumber, cblock.BlockNumber.N)
				issues = append(issues, issue)
			}
			cblock.BlockNumber.N = blockNumber
		} else if nblock, ok := newBlockMap[blockHash]; ok {
			if !nblock.BlockNumber.Equals2(blockNumber) {
				issue := db.NewDuplicatedBlockHashIssue(blockHash, blockNumber, nblock.BlockNumber.N)
				issues = append(issues, issue)
			}
			nblock.BlockNumber.N = blockNumber
		} else {
			newBlockMap[blockHash] = s.NewBlockHash(block)
		}
		for _, tx := range block.Transactions() {
			txHash := tx.Hash().Hex()
			if ctx, ok := changedTxMap[txHash]; ok {
				if !ctx.BlockNumber.Equals2(blockNumber) {
					issue := db.NewDuplicatedTxHashIssue(txHash, blockNumber, blockHash, ctx.BlockNumber.N, ctx.BlockHash)
					issues = append(issues, issue)
				}
				ctx.BlockNumber.N = blockNumber
				ctx.BlockHash = blockHash
			} else if ntx, ok := newTxMap[txHash]; ok {
				if !ntx.BlockNumber.Equals2(blockNumber) {
					issue := db.NewDuplicatedTxHashIssue(txHash, blockNumber, blockHash, ntx.BlockNumber.N, ntx.BlockHash)
					issues = append(issues, issue)
				}
				ntx.BlockNumber.N = blockNumber
				ntx.BlockHash = blockHash
			} else {
				newTxMap[txHash] = s.NewTxHash(block, tx)
			}
		}
	}
	for _, block := range newBlockMap {
		result.NewBlocks = append(result.NewBlocks, block)
	}
	for _, block := range changedBlockMap {
		result.ChangedBlocks = append(result.ChangedBlocks, block)
	}
	for _, tx := range newTxMap {
		result.NewTxs = append(result.NewTxs, tx)
	}
	for _, tx := range changedTxMap {
		result.ChangedTxs = append(result.ChangedTxs, tx)
	}
	result.Issues = issues
	return result, err
}

func (s *IndexBlockTxService) NewBlockHash(block *types.Block) *db.BlockHash {
	return &db.BlockHash{
		Hash:        block.Hash().Hex(),
		BlockNumber: &db.BigInt{N: block.Number()},
	}
}

func (s *IndexBlockTxService) NewTxHash(block *types.Block, tx *types.Transaction) *db.TxHash {
	return &db.TxHash{
		Hash:        tx.Hash().Hex(),
		BlockHash:   block.Hash().Hex(),
		BlockNumber: &db.BigInt{N: block.Number()},
	}
}

type GetBlockDataQueue struct {
	chanBlockNumbers   chan *GetBlockDataItem
	client             *rpc.EthClient
	logger             *zerolog.Logger
	BlockData          []*types.Block
	ChanCompleteSignal *sync.WaitGroup
	Error              error
}

type GetBlockDataItem struct {
	blockNumber *big.Int
	index       int
}

func (q *GetBlockDataQueue) Start() {
	for {
		select {
		case data := <-q.chanBlockNumbers:
			startTime := time.Now()
			var err error
			q.BlockData[data.index], err = q.client.GetBlockByNumber(data.blockNumber)
			if err != nil {
				q.BlockData[data.index] = nil
				q.Error = err
			}
			q.logger.Debug().Msgf("Fetched block #%d in %v", data.blockNumber.Int64(), time.Since(startTime))
			q.ChanCompleteSignal.Done()
		}
	}
}

func (q *GetBlockDataQueue) Equneue(blockNumber *big.Int, index int) {
	q.chanBlockNumbers <- &GetBlockDataItem{blockNumber, index}
}
