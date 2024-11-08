package rpc

import (
	"context"
	"math/big"
	"viction-rpc-crawler-go/x/ethutil"
)

func (client *EthClient) GetBlockFinalityByNumber(number *big.Int) (uint, error) {
	fn, err := rpcCall[uint](client, "eth_getBlockFinalityByNumber", ethutil.BigIntToHex(number))
	return *fn, err
}

func (client *EthClient) TraceBlockByNumber(number *big.Int) ([]*TraceTransactionResult, error) {
	tracerConfig := struct {
		Tracer  string `json:"tracer"`
		Timeout string `json:"timeout"`
	}{
		Tracer:  "callTracer",
		Timeout: "300s",
	}
	tempResult, err := rpcCall[[]TxTraceResult](client, "debug_traceBlockByNumber", ethutil.BigIntToHex(number), tracerConfig)
	if err != nil {
		return []*TraceTransactionResult{}, err
	}
	result := make([]*TraceTransactionResult, len(*tempResult))
	for i, r := range *tempResult {
		result[i] = r.Result
	}
	return result, err
}

func (client *EthClient) TraceTransaction(txHash string) (*TraceTransactionResult, error) {
	tracerConfig := struct {
		Tracer  string `json:"tracer"`
		Timeout string `json:"timeout"`
	}{
		Tracer:  "callTracer",
		Timeout: "300s",
	}
	result, err := rpcCall[TraceTransactionResult](client, "debug_traceTransaction", txHash, tracerConfig)
	return result, err
}

func rpcCall[T interface{}](client *EthClient, method string, args ...interface{}) (*T, error) {
	var result *T
	err := client.r.CallContext(context.Background(), &result, method, args...)
	return result, err
}
