package svc

import (
	"encoding/hex"
	"math/big"
	"slices"
	"sync"
	"time"

	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/x/ethutil"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gurukami/typ"
	"github.com/rs/zerolog"
)

var SYSTEM_ADDRESSES = []string{
	"0x0000000000000000000000000000000000000089", // Sign block
	"0x0000000000000000000000000000000000000090", // Randomize
	"0x0000000000000000000000000000000000000091", // TomoX
	"0x0000000000000000000000000000000000000092", // TomoXTradingState
	"0x0000000000000000000000000000000000000093", // TomoXLending
	"0x0000000000000000000000000000000000000094", // TomoXFinalLending
}

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
	startBlockNumber := big.NewInt(s.StartBlock)
	if s.UseHighestBlock {
		highestBlock, err := db.GetHighestIndexBlock()
		if err != nil {
			panic(err)
		}
		if highestBlock != nil {
			startBlockNumber = new(big.Int).SetUint64(highestBlock.BlockNumber)
		}
	}
	endBlock := big.NewInt(s.EndBlock)
	for startBlockNumber.Cmp(endBlock) <= 0 {
		endBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(int64(s.BatchSize)-1))
		if endBlockNumber.Cmp(endBlock) > 0 {
			endBlockNumber.Set(endBlock)
		}
		blocks, err := s.getBlockData(startBlockNumber, endBlockNumber)
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
		nextStartBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(int64(len(blocks))))
		err = db.SaveHighestIndexBlock(nextStartBlockNumber)
		if err != nil {
			panic(err)
		}
		s.Logger.Info().
			Int("NewBlocksCount", len(batchData.NewBlocks)).
			Int("ChangedBlocksCount", len(batchData.ChangedBlocks)).
			Int("NewTxsCount", len(batchData.NewTxs)).
			Int("ChangedTxsCount", len(batchData.ChangedTxs)).
			Int("IssuesCount", len(batchData.Issues)).
			Msgf("Persisted Batch #%d-%d in %v", startBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
		startBlockNumber.Set(nextStartBlockNumber)
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

func (s *IndexBlockTxService) getBlockData(startBlockNumber *big.Int, endBlockNumber *big.Int) ([]*rpc.Block, error) {
	batchSize := int(new(big.Int).Sub(endBlockNumber, startBlockNumber).Int64() + 1)
	s.workers.BlockData = make([]*rpc.Block, batchSize)
	s.workers.ChanCompleteSignal = new(sync.WaitGroup)
	s.workers.ChanCompleteSignal.Add(batchSize)
	startTime := time.Now()
	for i := 0; i < batchSize; i++ {
		s.workers.Equneue(new(big.Int).Add(startBlockNumber, big.NewInt(int64(i))), i)
	}
	s.workers.ChanCompleteSignal.Wait()
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

func (s *IndexBlockTxService) prepareBatchData(dbc *db.DbClient, blocks []*rpc.Block) (*IndexBlockBatchDataResult, error) {
	result := &IndexBlockBatchDataResult{
		NewBlocks:     []*db.Block{},
		ChangedBlocks: []*db.Block{},
		NewTxs:        []*db.Transaction{},
		ChangedTxs:    []*db.Transaction{},
		Issues:        []*db.Issue{},
	}

	blockHashes := []string{}
	txHashes := []string{}
	issues := []*db.Issue{}
	for _, block := range blocks {
		blockHashes = append(blockHashes, block.Hash.Hex())
		for _, tx := range block.Transactions {
			txHashes = append(txHashes, tx.Hash.Hex())
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
		changedTxMap[tx.Hash] = tx
	}
	for _, block := range blocks {
		blockNumber := block.Number.BigInt()
		blockHash := block.Hash.Hex()
		systemTxCount := uint16(0)
		for _, tx := range block.Transactions {
			txHash := tx.Hash.Hex()
			toAddress := ""
			if tx.To != nil {
				toAddress = tx.To.Hex()
			}
			if slices.Contains(SYSTEM_ADDRESSES, toAddress) {
				systemTxCount += 1
			}
			if ctx, ok := changedTxMap[txHash]; ok {
				if ctx.BlockID != blockNumber.Uint64() {
					issue := db.NewDuplicatedTxHashIssue(tx.Hash.Hex(), blockNumber.Uint64(), blockHash, ctx.BlockID, ctx.BlockHash)
					issues = append(issues, issue)
				}
				s.copyTransactionProperties(tx, block, ctx)
			} else if ntx, ok := newTxMap[txHash]; ok {
				if ntx.BlockID != blockNumber.Uint64() {
					issue := db.NewDuplicatedTxHashIssue(tx.Hash.Hex(), blockNumber.Uint64(), blockHash, ntx.BlockID, ntx.BlockHash)
					issues = append(issues, issue)
				}
				s.copyTransactionProperties(tx, block, ntx)
			} else {
				ntx := &db.Transaction{Hash: tx.Hash.Hex()}
				s.copyTransactionProperties(tx, block, ntx)
				newTxMap[txHash] = ntx
			}
		}

		if cblock, ok := changedBlockMap[blockNumber.Uint64()]; ok {
			if cblock.Hash != blockHash {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
			}
			s.copyBlockProperties(block, cblock)
			cblock.TransactionCountSystem.Scan(&systemTxCount)
		} else if nblock, ok := newBlockMap[blockNumber.Uint64()]; ok {
			if nblock.Hash != blockHash {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
			}
			s.copyBlockProperties(block, nblock)
			nblock.TransactionCountSystem.Scan(&systemTxCount)
		} else {
			nblock := &db.Block{ID: blockNumber.Uint64()}
			s.copyBlockProperties(block, nblock)
			newBlockMap[blockNumber.Uint64()] = nblock
			nblock.TransactionCountSystem.Scan(&systemTxCount)
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

func (s *IndexBlockTxService) copyBlockProperties(ethBlock *rpc.Block, dbBlock *db.Block) {
	dbBlock.Hash = ethBlock.Hash.Hex()
	dbBlock.Timestamp = int64(ethBlock.Timestamp.Int())
	dbBlock.Size = uint16(ethBlock.Size.Int())
	dbBlock.GasLimit = ethBlock.GasLimit.Int()
	dbBlock.GasUsed = ethBlock.GasUsed.Int()
	dbBlock.Difficulty = ethBlock.Difficulty.Decimal()
	dbBlock.TotalDifficulty = ethBlock.TotalDifficulty.Decimal()
	dbBlock.TransactionCount = uint16(len(ethBlock.Transactions))
	dbBlock.TransactionCountSystem = typ.NullUint16{}
	dbBlock.TransactionCountDebug = typ.NullUint16{}
	dbBlock.BlockMintDuration = typ.NullUint64{}
	dbBlock.ParentHash = ethBlock.ParentHash.Hex()
	dbBlock.UncleHash = ethBlock.Sha3Uncles.Hex()
	dbBlock.StateRoot = ethBlock.StateRoot.Hex()
	dbBlock.TransactionsRoot = ethBlock.TransactionsRoot.Hex()
	dbBlock.ReceiptsRoot = ethBlock.ReceiptsRoot.Hex()
	dbBlock.LogsBloom = ethBlock.LogsBloom.Hex()
	dbBlock.Miner = ethBlock.Miner.Hex()
	dbBlock.ExtraData = ethBlock.ExtraData.Hex()
	dbBlock.MixDigest = ethBlock.MixDigest.Hex()
	dbBlock.Nonce = ethBlock.Nonce.Hex()
	dbBlock.Validator = ethBlock.Validator.Hex()
	dbBlock.Creator = typ.NullString{}
	dbBlock.Attestor = typ.NullString{}
	signatureLength := 65
	extraData := ethBlock.ExtraData.Bytes()
	creatorSignature := extraData[len(extraData)-signatureLength:]
	creator, err := crypto.Ecrecover(ethBlock.SigHash(), creatorSignature)
	if err == nil {
		addr := ethutil.PubkeyToAddress(creator)
		dbBlock.Creator.Set(hex.EncodeToString(addr))
	}
	if ethBlock.Validator != nil && len(ethBlock.Validator.Bytes()) >= signatureLength {
		validatorBytes := ethBlock.Validator.Bytes()
		attestorSignature := validatorBytes[len(validatorBytes)-signatureLength:]
		attestor, err := crypto.Ecrecover(ethBlock.SigHash(), attestorSignature)
		if err == nil {
			addr := ethutil.PubkeyToAddress(attestor)
			dbBlock.Attestor.Set(hex.EncodeToString(addr))
		}
	}
}

func (s *IndexBlockTxService) copyTransactionProperties(ethTransaction *rpc.Transaction, ethBlock *rpc.Block, dbTransaction *db.Transaction) {
	dbTransaction.BlockID = ethBlock.Number.Int()
	dbTransaction.BlockHash = ethBlock.Hash.Hex()
	dbTransaction.TransactionIndex = 0
	dbTransaction.From = ethTransaction.From.Hex()
	if ethTransaction.To != nil {
		dbTransaction.To = ethTransaction.To.Hex()
	}
	dbTransaction.Value = ethTransaction.Value.Decimal()
	dbTransaction.Nonce = ethTransaction.Nonce.Int()
	dbTransaction.Gas = ethTransaction.Gas.Int()
	dbTransaction.GasPrice = ethTransaction.GasPrice.Decimal()
}

type GetBlockDataQueue struct {
	chanBlockNumbers   chan *GetBlockDataItem
	client             *rpc.EthClient
	logger             *zerolog.Logger
	BlockData          []*rpc.Block
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
			q.BlockData[data.index], err = q.client.GetBlockByNumber2(data.blockNumber)
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
