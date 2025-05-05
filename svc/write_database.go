package svc

import (
	"encoding/hex"
	"slices"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/ethutil"
	"viction-rpc-crawler-go/rpc"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gurukami/typ"
	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

var SystemAddresses = []string{
	"0000000000000000000000000000000000000089", // Sign block
	"0000000000000000000000000000000000000090", // Randomize
	"0000000000000000000000000000000000000091", // TomoX
	"0000000000000000000000000000000000000092", // TomoXTradingState
	"0000000000000000000000000000000000000093", // TomoXLending
	"0000000000000000000000000000000000000094", // TomoXFinalLending
}

type WriteDatabase struct {
	multiplex.ServiceCore
	i  *multiplex.ServiceCoreInternal
	db *db.DbClient
}

func NewWriteDatabase(logger diag.Logger, dbClient *db.DbClient) *WriteDatabase {
	svc := &WriteDatabase{
		db: dbClient,
	}
	svc.i = svc.InitServiceCore("WriteDatabase", logger, svc.coreProcessHook)
	return svc
}

func (s *WriteDatabase) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "write_blocks":
		blocks := msg.GetParam("blocks", []*rpc.Block{}).([]*rpc.Block)
		batchData, err := s.prepareBatchData(blocks)
		if err != nil {
			panic(err)
		}
		if len(batchData.NewBlocks)+len(batchData.ChangedBlocks) > 0 {
			err = s.db.SaveBlocks(batchData.NewBlocks, batchData.ChangedBlocks)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.NewTxs)+len(batchData.ChangedTxs) > 0 {
			err = s.db.SaveTransactions(batchData.NewTxs, batchData.ChangedTxs)
			if err != nil {
				panic(err)
			}
		}
		if len(batchData.Issues) > 0 {
			err = s.db.SaveIssues(batchData.Issues)
			if err != nil {
				panic(err)
			}
		}
		if err != nil {
			panic(err)
		}
		defer msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}

func (s *WriteDatabase) prepareBatchData(blocks []*rpc.Block) (*BlockBatchData, error) {
	result := &BlockBatchData{
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

	newBlockMap := make(map[uint64]*db.Block)
	changedBlockMap := make(map[uint64]*db.Block)
	changedBlocks, err := s.db.GetBlocksByHashes(blockHashes)
	if err != nil {
		return nil, err
	}
	for _, block := range changedBlocks {
		changedBlockMap[block.ID] = block
	}
	newTxMap := make(map[string]*db.Transaction)
	changedTxMap := make(map[string]*db.Transaction)
	changedTxs, err := s.db.GetTransactions(txHashes)
	if err != nil {
		return nil, err
	}
	for _, tx := range changedTxs {
		changedTxMap[tx.Hash] = tx
	}
	for _, block := range blocks {
		blockNumber := block.Number.BigInt()
		blockHash := block.Hash.Hex()
		txCount := uint16(len(block.Transactions))
		systemTxCount := uint16(0)
		for _, tx := range block.Transactions {
			txHash := tx.Hash.Hex()
			toAddress := ""
			if tx.To != nil {
				toAddress = tx.To.Hex()
			}
			if slices.Contains(SystemAddresses, toAddress) {
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
			cblock.TransactionCount.Scan(&txCount)
			cblock.TransactionCountSystem.Scan(&systemTxCount)
		} else if nblock, ok := newBlockMap[blockNumber.Uint64()]; ok {
			if nblock.Hash != blockHash {
				issue := db.NewReorgBlockIssue(blockNumber.Uint64(), cblock.Hash, blockHash)
				issues = append(issues, issue)
			}
			s.copyBlockProperties(block, nblock)
			nblock.TransactionCount.Scan(&txCount)
			nblock.TransactionCountSystem.Scan(&systemTxCount)
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
	for _, tx := range newTxMap {
		result.NewTxs = append(result.NewTxs, tx)
	}
	for _, tx := range changedTxMap {
		result.ChangedTxs = append(result.ChangedTxs, tx)
	}
	result.Issues = issues
	return result, err
}

func (s *WriteDatabase) copyBlockProperties(ethBlock *rpc.Block, dbBlock *db.Block) {
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

func (s *WriteDatabase) copyTransactionProperties(ethTransaction *rpc.Transaction, ethBlock *rpc.Block, dbTransaction *db.Transaction) {
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

type BlockBatchData struct {
	NewBlocks     []*db.Block
	ChangedBlocks []*db.Block
	NewTxs        []*db.Transaction
	ChangedTxs    []*db.Transaction
	Issues        []*db.Issue
}
