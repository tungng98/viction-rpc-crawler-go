package svc

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"sync"
	"time"
	"viction-rpc-crawler-go/cache"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/x/ethutil"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gurukami/typ"
	"github.com/rs/zerolog"
)

type BlockIndexerCommand struct {
	Command string
	Params  ExecParams
}

type BlockIndexerBatchDataResult struct {
	NewBlocks        []*db.Block
	ChangedBlocks    []*db.Block
	TransactionCount int
	Issues           []*db.Issue
	Reorg            bool
	Errors           []*GetBlockResult
}

type BlockIndexerSvc struct {
	i *BlockIndexerSvcInternal
}

type BlockIndexerSvcInternal struct {
	WorkerMainCounter  *WorkerCounter
	WorkerMainCount    uint16
	WorkerID           uint64
	IndexerMainCounter *WorkerCounter

	MainChan chan *BlockIndexerCommand
	MainWait map[string]*sync.WaitGroup

	Controller  ServiceController
	Rpc         *rpc.EthClient
	SharedCache cache.MemoryCache
	Db          *db.DbClient
	Logger      *zerolog.Logger
}

func NewBlockIndexerSvc(controller ServiceController, rpc *rpc.EthClient, cache cache.MemoryCache, db *db.DbClient, logger *zerolog.Logger) BlockIndexerSvc {
	svc := BlockIndexerSvc{
		i: &BlockIndexerSvcInternal{
			WorkerMainCounter:  &WorkerCounter{},
			IndexerMainCounter: &WorkerCounter{},

			MainChan: make(chan *BlockIndexerCommand, MAIN_CHAN_CAPACITY),
			MainWait: map[string]*sync.WaitGroup{},

			Controller:  controller,
			Rpc:         rpc,
			SharedCache: cache,
			Db:          db,
			Logger:      logger,
		},
	}
	return svc
}

func (s BlockIndexerSvc) ServiceID() string {
	return "BlockIndexer"
}

func (s BlockIndexerSvc) Controller() ServiceController {
	return s.i.Controller
}

func (s BlockIndexerSvc) SetWorker(workerCount uint16) {
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
			cmd := &BlockIndexerCommand{
				Command: "exit",
			}
			s.i.MainChan <- cmd
		}
		s.i.WorkerMainCount = workerCount
	}
}

func (s BlockIndexerSvc) WorkerCount() uint16 {
	return s.i.WorkerMainCounter.ValueNoLock()
}

func (s BlockIndexerSvc) Exec(command string, params ExecParams) {
	if command == "exit" {
		return
	}
	if command == "index" {
		if s.i.IndexerMainCounter.Value() == 0 {
			msg := &BlockIndexerCommand{
				Command: command,
				Params:  params,
			}
			s.i.MainChan <- msg
		} else {
			s.i.Logger.Info().Msg("Main indexer is running. Exec ignored.")
		}
	}
}

func (s *BlockIndexerSvc) process(workerID uint64) {
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockIndexer Process started.")
	s.i.WorkerMainCounter.Increase()
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.MainChan
		switch msg.Command {
		case "index":
			s.i.IndexerMainCounter.Increase()
			from := msg.Params.Get("from", big.NewInt(1)).(*big.Int)
			to := msg.Params.Get("to", big.NewInt(999999999)).(*big.Int)
			batchSize := msg.Params.Get("batch_size", 1).(int)
			useCheckpointBlock := msg.Params.Get("use_checkpoint_block", true).(bool)
			s.indexBlock(from, to, batchSize, useCheckpointBlock)
			s.i.IndexerMainCounter.Decrease()
		case "exit":
			status = EXIT_STATE
		}
	}
	s.i.WorkerMainCounter.Decrease()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("BlockIndexer Process exited.")
}

func (s *BlockIndexerSvc) indexBlock(from, to *big.Int, batchSize int, useCheckpointBlock bool) {
	startBlockNumber := new(big.Int).Set(from)
	if useCheckpointBlock {
		highestBlock, err := s.i.Db.GetHighestIndexBlock()
		if err != nil {
			panic(err)
		}
		if highestBlock != nil {
			startBlockNumber = new(big.Int).SetUint64(highestBlock.BlockNumber)
		}
	}
	endBlock := new(big.Int).Set(to)
	head, err := s.i.Rpc.GetBlockNumber()
	if err != nil {
		panic(err)
	}
	if to.Uint64() > head {
		endBlock.SetUint64(head)
	}
	for startBlockNumber.Cmp(endBlock) < 0 {
		endBlockRounding := true
		endBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(int64(batchSize)))
		if endBlockNumber.Cmp(endBlock) > 0 {
			endBlockNumber.Set(endBlock)
			endBlockRounding = false
		}
		if endBlockRounding {
			endBlockNumberMod := new(big.Int).Mod(endBlockNumber, big.NewInt(int64(batchSize)))
			if endBlockNumberMod.Cmp(big.NewInt(0)) != 0 {
				endBlockNumber = new(big.Int).Sub(endBlockNumber, endBlockNumberMod)
			}
		}

		requestId := fmt.Sprintf("get_blocks_range_%d_%d", startBlockNumber.Uint64(), endBlockNumber.Uint64())
		getBlockDataParams := ExecParams{
			"request_id": requestId,
			"from":       startBlockNumber,
			"to":         endBlockNumber,
		}
		getBlockDataParams.ExpectReturns()
		s.i.Controller.ExecService("BlockFetcher", "get_blocks_range", getBlockDataParams)
		getBlockDataParams.WaitForReturns()

		getBlockResults := cache.GetArray[*GetBlockResult](s.i.SharedCache, requestId)
		cache.DeleteArray(s.i.SharedCache, requestId)
		startTime := time.Now()
		batchData, err := s.prepareBatchData(s.i.Db, getBlockResults)
		if err != nil {
			panic(err)
		}
		if batchData.Reorg {
			startBlockNumber = new(big.Int).Sub(startBlockNumber, big.NewInt(10))
			continue
		}
		if len(batchData.NewBlocks)+len(batchData.ChangedBlocks) > 0 {
			err = s.i.Db.SaveBlocks(batchData.NewBlocks, batchData.ChangedBlocks)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.Issues) > 0 {
			err = s.i.Db.SaveIssues(batchData.Issues)
			if err != nil {
				panic(err)
			}
		}
		if useCheckpointBlock {
			err = s.i.Db.SaveHighestIndexBlock(endBlockNumber)
			if err != nil {
				panic(err)
			}
		}
		s.i.Logger.Info().
			Int("NewBlocksCount", len(batchData.NewBlocks)).
			Int("ChangedBlocksCount", len(batchData.ChangedBlocks)).
			Int("TxsCount", batchData.TransactionCount).
			Int("IssuesCount", len(batchData.Issues)).
			Msgf("Persisted Batch #%d-%d in %v.", startBlockNumber.Int64(), endBlockNumber.Int64(), time.Since(startTime))
		startBlockNumber.Set(endBlockNumber)
	}
}

func (s *BlockIndexerSvc) prepareBatchData(dbc *db.DbClient, getBlockResults []*GetBlockResult) (*BlockIndexerBatchDataResult, error) {
	result := &BlockIndexerBatchDataResult{
		NewBlocks:     []*db.Block{},
		ChangedBlocks: []*db.Block{},
		Issues:        []*db.Issue{},
		Errors:        []*GetBlockResult{},
	}

	blockIDs := []uint64{}
	issues := []*db.Issue{}
	for _, getBlock := range getBlockResults {
		if getBlock.Error != nil {
			result.Errors = append(result.Errors, getBlock)
			continue
		}
		blockIDs = append(blockIDs, getBlock.BlockNumber.Uint64())
	}

	newBlockMap := make(map[uint64]*db.Block)
	changedBlockMap := make(map[uint64]*db.Block)
	changedBlocks, err := dbc.GetBlocks(blockIDs)
	if err != nil {
		return nil, err
	}
	for _, block := range changedBlocks {
		changedBlockMap[block.ID] = block
	}
	for i, getBlock := range getBlockResults {
		block := getBlock.Data
		blockNumber := block.Number.BigInt()
		blockHash := block.Hash.Hex()
		if i == 0 {
			if cblock, ok := changedBlockMap[blockNumber.Uint64()]; ok {
				if cblock.Hash != blockHash {
					result.Reorg = true
				}
			}
		}
		txCount := uint16(len(block.Transactions))
		result.TransactionCount += int(txCount)
		systemTxCount := uint16(0)
		for _, tx := range block.Transactions {
			toAddress := ""
			if tx.To != nil {
				toAddress = tx.To.Hex()
			}
			if slices.Contains(SYSTEM_ADDRESSES, toAddress) {
				systemTxCount += 1
			}
		}

		if cblock, ok := changedBlockMap[blockNumber.Uint64()]; ok {
			if cblock.Hash == blockHash {
				delete(changedBlockMap, blockNumber.Uint64())
			} else {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
				s.copyBlockProperties(block, cblock)
				cblock.TransactionCount.Scan(&txCount)
				cblock.TransactionCountSystem.Scan(&systemTxCount)
			}
		} else {
			nblock := &db.Block{ID: blockNumber.Uint64()}
			s.copyBlockProperties(block, nblock)
			newBlockMap[blockNumber.Uint64()] = nblock
			nblock.TransactionCount.Scan(&txCount)
			nblock.TransactionCountSystem.Scan(&systemTxCount)
		}
	}
	for _, block := range newBlockMap {
		result.NewBlocks = append(result.NewBlocks, block)
	}
	for _, block := range changedBlockMap {
		result.ChangedBlocks = append(result.ChangedBlocks, block)
	}
	result.Issues = issues
	return result, err
}

func (s *BlockIndexerSvc) copyBlockProperties(ethBlock *rpc.Block, dbBlock *db.Block) {
	dbBlock.Hash = ethBlock.Hash.Hex()
	dbBlock.ParentHash = ethBlock.ParentHash.Hex()
	dbBlock.Timestamp = int64(ethBlock.Timestamp.Int())
	dbBlock.Size = uint16(ethBlock.Size.Int())
	dbBlock.GasLimit = ethBlock.GasLimit.Int()
	dbBlock.GasUsed = ethBlock.GasUsed.Int()
	dbBlock.Difficulty = ethBlock.Difficulty.Decimal()
	dbBlock.TotalDifficulty = ethBlock.TotalDifficulty.Decimal()
	dbBlock.TransactionCount = typ.NullUint16{}
	dbBlock.TransactionCountSystem = typ.NullUint16{}
	dbBlock.TransactionCountDebug = typ.NullUint16{}
	dbBlock.BlockMintDuration = typ.NullUint64{}
	dbBlock.UncleHash = ethBlock.Sha3Uncles.Bytes()
	dbBlock.StateRoot = ethBlock.StateRoot.Bytes()
	dbBlock.TransactionsRoot = ethBlock.TransactionsRoot.Bytes()
	dbBlock.ReceiptsRoot = ethBlock.ReceiptsRoot.Bytes()
	dbBlock.LogsBloom = ethBlock.LogsBloom.Bytes()
	dbBlock.Miner = ethBlock.Miner.Bytes()
	dbBlock.ExtraData = ethBlock.ExtraData.Bytes()
	dbBlock.MixDigest = ethBlock.MixDigest.Bytes()
	dbBlock.Nonce = ethBlock.Nonce.Bytes()
	dbBlock.Validator = ethBlock.Validator.Bytes()
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
