package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"viction-rpc-crawler-go/x/ethutil"
)

func (client *EthClient) GetBlockFinalityByNumber(number *big.Int) (uint, error) {
	fn, err := rpcCall[uint](client, "eth_getBlockFinalityByNumber", ethutil.BigIntToHex(number))
	return *fn, err
}

func (client *EthClient) TraceTransaction(txHash string) (*json.RawMessage, error) {
	tracerConfig := struct {
		Tracer  string `json:"tracer"`
		Timeout string `json:"timeout"`
	}{
		Tracer:  "callTracer",
		Timeout: "300s",
	}
	fn, err := rpcCall[json.RawMessage](client, "debug_traceTransaction", txHash, tracerConfig)
	return fn, err
}

func rpcCall[T interface{}](client *EthClient, method string, args ...interface{}) (*T, error) {
	var result *T
	err := client.r.CallContext(context.Background(), &result, method, args...)
	if err != nil {
		return nil, err
	}
	return result, err
}
