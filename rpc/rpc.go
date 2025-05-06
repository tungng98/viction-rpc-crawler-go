package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"viction-rpc-crawler-go/ethutil"
)

func (client *EthClient) GetBlockByNumber2(number *big.Int) (*Block, string, error) {
	fn, str, err := rpcCall[Block](client, "eth_getBlockByNumber", ethutil.BigIntToHex(number), true)
	return fn, str, err
}

func (client *EthClient) GetBlockFinalityByNumber(number *big.Int) (*uint, string, error) {
	fn, str, err := rpcCall[uint](client, "eth_getBlockFinalityByNumber", ethutil.BigIntToHex(number))
	return fn, str, err
}

func (client *EthClient) TraceBlockByNumber(number *big.Int) (TraceBlockResult, string, error) {
	tracerConfig := struct {
		Tracer  string `json:"tracer"`
		Timeout string `json:"timeout"`
	}{
		Tracer:  "callTracer",
		Timeout: "300s",
	}
	tempResult, str, err := rpcCall[[]TxTraceResult](client, "debug_traceBlockByNumber", ethutil.BigIntToHex(number), tracerConfig)
	if err != nil {
		return TraceBlockResult{}, str, err
	}
	result := make(TraceBlockResult, len(*tempResult))
	for i, r := range *tempResult {
		result[i] = r.Result
	}
	return result, str, err
}

func (client *EthClient) TraceTransaction(txHash string) (*TraceTransactionResult, string, error) {
	tracerConfig := struct {
		Tracer  string `json:"tracer"`
		Timeout string `json:"timeout"`
	}{
		Tracer:  "callTracer",
		Timeout: "300s",
	}
	result, str, err := rpcCall[TraceTransactionResult](client, "debug_traceTransaction", txHash, tracerConfig)
	return result, str, err
}

func rpcCall[T interface{}](client *EthClient, method string, args ...interface{}) (*T, string, error) {
	var raw json.RawMessage
	err := client.r.CallContext(context.Background(), &raw, method, args...)
	var result *T
	if err == nil && raw != nil {
		err = json.Unmarshal([]byte(raw), &result)
	}
	return result, string(raw), err
}
