package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"viction-rpc-crawler-go/ethutil"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
)

type Block struct {
	Number           *Uint256 `json:"number,omitempty"`
	Hash             *Hex     `json:"hash,omitempty"`
	Timestamp        *Uint256 `json:"timestamp,omitempty"`
	Size             *Uint64  `json:"size,omitempty"`
	GasLimit         *Uint64  `json:"gasLimit,omitempty"`
	GasUsed          *Uint64  `json:"gasUsed,omitempty"`
	Difficulty       *Uint256 `json:"difficulty,omitempty"`
	TotalDifficulty  *Uint256 `json:"totalDifficulty,omitempty"`
	Nonce            *Hex     `json:"nonce,omitempty"`
	ExtraData        *Hex     `json:"extraData,omitempty"`
	LogsBloom        *Hex     `json:"logsBloom,omitempty"`
	ParentHash       *Hex     `json:"parentHash,omitempty"`
	StateRoot        *Hex     `json:"stateRoot,omitempty"`
	TransactionsRoot *Hex     `json:"transactionsRoot,omitempty"`
	ReceiptsRoot     *Hex     `json:"receiptsRoot,omitempty"`
	Sha3Uncles       *Hex     `json:"sha3Uncles,omitempty"`
	MixDigest        *Hex     `json:"mixHash,omitempty"`
	Miner            *Hex     `json:"miner,omitempty"`
	Validator        *Hex     `json:"validator,omitempty"`
	Validators       *Hex     `json:"validators,omitempty"`
	Penalties        *Hex     `json:"penalties,omitempty"`

	Transactions []*Transaction `json:"transactions,omitempty"`
}

func (b *Block) SigHash() []byte {
	hasher := sha3.NewLegacyKeccak256()
	hash := make([]byte, 32)

	rlp.Encode(hasher, []interface{}{
		b.ParentHash.Bytes(),
		b.Sha3Uncles.Bytes(),
		b.Miner.Bytes(),
		b.StateRoot.Bytes(),
		b.TransactionsRoot.Bytes(),
		b.ReceiptsRoot.Bytes(),
		b.LogsBloom.Bytes(),
		b.Difficulty.BigInt(),
		b.Number.BigInt(),
		b.GasLimit.Int(),
		b.GasUsed.Int(),
		b.Timestamp.BigInt(),
		b.ExtraData.Bytes()[:len(b.ExtraData.Bytes())-65], // Yes, this will panic if extra is too short
		b.MixDigest.Bytes(),
		b.Nonce.Bytes(),
	})
	hasher.Sum(hash[:0])
	return hash
}

type Transaction struct {
	Hash        *Hex     `json:"hash,omitempty"`
	BlockNumber *Uint256 `json:"blockNumber,omitempty"`
	BlockHash   *Hex     `json:"blockHash,omitempty"`
	From        *Hex     `json:"from,omitempty"`
	To          *Hex     `json:"to,omitempty"`
	Value       *Uint256 `json:"value,omitempty"`
	Input       *Hex     `json:"input,omitempty"`
	Gas         *Uint256 `json:"gas,omitempty"`
	GasPrice    *Uint256 `json:"gasPrice,omitempty"`
	Nonce       *Uint64  `json:"nonce,omitempty"`
	Index       *Uint64  `json:"transactionIndex,omitempty"`
	V           *Hex     `json:"v,omitempty"`
	R           *Hex     `json:"r,omitempty"`
	S           *Hex     `json:"s,omitempty"`
}

type TxTraceResult struct {
	TxHash string                  `json:"txHash,omitempty"`
	Result *TraceTransactionResult `json:"result,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

type TraceBlockResult []*TraceTransactionResult

type TraceTransactionResult struct {
	Type    string `json:"type,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Value   string `json:"value,omitempty"`
	Gas     string `json:"gas,omitempty"`
	GasUsed string `json:"gasUsed,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`
	Time    string `json:"time,omitempty"`

	Calls []*TraceTransactionCall `json:"calls,omitempty"`
}

type TraceTransactionCall struct {
	Type    string `json:"type,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Value   string `json:"value,omitempty"`
	Gas     string `json:"gas,omitempty"`
	GasUsed string `json:"gasUsed,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`

	Calls []*TraceTransactionCall `json:"calls,omitempty"`
}

type Hex struct {
	i []byte
}

func (h *Hex) Bytes() []byte {
	return h.i
}

func (h *Hex) Hex() string {
	return hex.EncodeToString(h.i)
}

func (h *Hex) Hex0x() string {
	return "0x" + hex.EncodeToString(h.i)
}

func (h Hex) MarshalJSON() ([]byte, error) {
	hex := h.Hex0x()
	return json.Marshal(hex)
}

func (h *Hex) UnmarshalJSON(data []byte) error {
	hex := string(data)
	hex = strings.TrimPrefix(hex, "\"")
	hex = strings.TrimSuffix(hex, "\"")

	matched, err := regexp.Match("^0x[0-9a-f]*$", []byte(hex))
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid hex input: %s", hex)
	}

	h.i = ethutil.HexToBytes(hex)
	return nil
}

type Uint64 struct {
	i uint64
}

func (n *Uint64) Int() uint64 {
	return n.i
}

func (n Uint64) MarshalJSON() ([]byte, error) {
	uintBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uintBytes, n.i)
	hex := "0x" + hex.EncodeToString(uintBytes)
	return json.Marshal(hex)
}

func (n *Uint64) UnmarshalJSON(data []byte) error {
	hex := string(data)
	hex = strings.TrimPrefix(hex, "\"")
	hex = strings.TrimSuffix(hex, "\"")

	matched, err := regexp.MatchString(`(?sm)^0x[0-9a-f]{0,16}$`, hex)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid hex input: %s", hex)
	}

	if bi, ok := new(big.Int).SetString(hex, 0); ok {
		n.i = bi.Uint64()
	} else {
		return fmt.Errorf("invalid hex input: %s", hex)
	}
	return nil
}

type Uint256 struct {
	i *big.Int
}

func (n *Uint256) Int() uint64 {
	return n.i.Uint64()
}

func (n *Uint256) BigInt() *big.Int {
	return n.i
}

func (n *Uint256) Decimal() decimal.Decimal {
	return decimal.NewFromBigInt(n.i, 0)
}

func (n Uint256) MarshalJSON() ([]byte, error) {
	hex := ethutil.BigIntToHex(n.i)
	return json.Marshal(hex)
}

func (n *Uint256) UnmarshalJSON(data []byte) error {
	hex := string(data)
	hex = strings.TrimPrefix(hex, "\"")
	hex = strings.TrimSuffix(hex, "\"")

	matched, err := regexp.MatchString(`(?m)^0x[0-9a-f]{0,64}`, hex)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid hex input: %s", hex)
	}

	if bi, ok := new(big.Int).SetString(hex, 0); ok {
		n.i = bi
	} else {
		return fmt.Errorf("invalid hex input: %s", hex)
	}
	return nil
}
