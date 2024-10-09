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

func rpcCall[T interface{}](client *EthClient, method string, args ...interface{}) (*T, error) {
	var result *T
	err := client.r.CallContext(context.Background(), &result, method, args...)
	if err != nil {
		return nil, err
	}
	return result, err
}
