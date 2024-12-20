package svc

import (
	"encoding/hex"
	"math/big"
	"sync"
	"time"

	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/x/ethutil"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/tee8z/nullable"
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
			err = db.SaveBlocks(batchData.NewBlocks, batchData.ChangedBlocks)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.NewTxs)+len(batchData.ChangedTxs) > 0 {
			err = db.SaveTransactions(batchData.NewTxs, batchData.ChangedTxs)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.Issues) > 0 {
			err = db.SaveIssues(batchData.Issues)
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
	NewBlocks     []*db.Block
	ChangedBlocks []*db.Block
	NewTxs        []*db.Transaction
	ChangedTxs    []*db.Transaction
	Issues        []*db.Issue
}

func (s *IndexBlockTxService) prepareBatchData(dbc *db.DbClient, blocks []*types.Block) (*IndexBlockBatchDataResult, error) {
	result := &IndexBlockBatchDataResult{
		NewBlocks:     []*db.Block{},
		ChangedBlocks: []*db.Block{},
		NewTxs:        []*db.Transaction{},
		ChangedTxs:    []*db.Transaction{},
		Issues:        []*db.Issue{},
	}

	blockHashes := [][]byte{}
	txHashes := [][]byte{}
	issues := []*db.Issue{}
	for _, block := range blocks {
		blockHashes = append(blockHashes, block.Hash().Bytes())
		for _, tx := range block.Transactions() {
			txHashes = append(txHashes, tx.Hash().Bytes())
		}
	}

	changedBlocks, err := dbc.GetBlocksByHashes(blockHashes)
	if err != nil {
		return nil, err
	}
	changedTxs, err := dbc.GetTransactions(txHashes)
	if err != nil {
		return nil, err
	}
	newBlockMap := make(map[uint64]*db.Block)
	changedBlockMap := make(map[uint64]*db.Block)
	newTxMap := make(map[string]*db.Transaction)
	changedTxMap := make(map[string]*db.Transaction)
	for _, block := range changedBlocks {
		changedBlockMap[block.ID] = block
	}
	for _, tx := range changedTxs {
		changedTxMap[hex.EncodeToString(tx.Hash)] = tx
	}
	for _, block := range blocks {
		blockNumber := block.Number()
		blockHash := block.Hash().Bytes()
		if cblock, ok := changedBlockMap[blockNumber.Uint64()]; ok {
			if !ethutil.BytesEqual(cblock.Hash, blockHash) {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
			}
			s.copyBlockProperties(block, cblock)
		} else if nblock, ok := newBlockMap[blockNumber.Uint64()]; ok {
			if !ethutil.BytesEqual(nblock.Hash, blockHash) {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
			}
			s.copyBlockProperties(block, nblock)
		} else {
			nblock := &db.Block{ID: blockNumber.Uint64()}
			s.copyBlockProperties(block, nblock)
			newBlockMap[blockNumber.Uint64()] = nblock
		}
		for _, tx := range block.Transactions() {
			txHash := tx.Hash().Hex()
			if ctx, ok := changedTxMap[txHash]; ok {
				if ctx.BlockID != blockNumber.Uint64() {
					issue := db.NewDuplicatedTxHashIssue(tx.Hash().Bytes(), blockNumber.Uint64(), blockHash, ctx.BlockID, ctx.BlockHash)
					issues = append(issues, issue)
				}
				s.copyTransactionProperties(tx, block, ctx)
			} else if ntx, ok := newTxMap[txHash]; ok {
				if ntx.BlockID != blockNumber.Uint64() {
					issue := db.NewDuplicatedTxHashIssue(tx.Hash().Bytes(), blockNumber.Uint64(), blockHash, ntx.BlockID, ntx.BlockHash)
					issues = append(issues, issue)
				}
				s.copyTransactionProperties(tx, block, ntx)
			} else {
				ntx := &db.Transaction{Hash: tx.Hash().Bytes()}
				s.copyTransactionProperties(tx, block, ntx)
				newTxMap[txHash] = ntx
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

func (s *IndexBlockTxService) copyBlockProperties(ethBlock *types.Block, dbBlock *db.Block) {
	dbBlock.Hash = ethBlock.Hash().Bytes()
	dbBlock.ParentHash = ethBlock.ParentHash().Bytes()
	dbBlock.StateRoot = ethBlock.Root().Bytes()
	dbBlock.TransactionsRoot = ethBlock.TxHash().Bytes()
	dbBlock.ReceiptsRoot = ethBlock.ReceiptHash().Bytes()
	dbBlock.Timestamp = int64(ethBlock.Header().Time)
	dbBlock.Size = uint16(ethBlock.Size())
	dbBlock.GasLimit = ethBlock.GasLimit()
	dbBlock.GasUsed = ethBlock.GasUsed()
	dbBlock.TotalDifficulty = ethBlock.Difficulty().Uint64()
	dbBlock.TransactionCount = uint16(len(ethBlock.Transactions()))
	dbBlock.TransactionCountSystem = nullable.NewUint16(nil)
	dbBlock.TransactionCountDebug = nullable.NewUint16(nil)
	dbBlock.BlockMintDuration = nullable.NewUint64(nil)
}

func (s *IndexBlockTxService) copyTransactionProperties(ethTransaction *types.Transaction, ethBlock *types.Block, dbTransaction *db.Transaction) {
	from, _ := types.Sender(types.NewEIP155Signer(ethTransaction.ChainId()), ethTransaction)
	dbTransaction.BlockID = ethBlock.Number().Uint64()
	dbTransaction.BlockHash = ethBlock.Hash().Bytes()
	dbTransaction.TransactionIndex = 0
	dbTransaction.From = from.Bytes()
	dbTransaction.To = ethTransaction.To().Bytes()
	dbTransaction.Value = ethTransaction.Value().Uint64()
	dbTransaction.Nonce = ethTransaction.Nonce()
	dbTransaction.Gas = ethTransaction.Gas()
	dbTransaction.GasPrice = ethTransaction.GasPrice().Uint64()
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
