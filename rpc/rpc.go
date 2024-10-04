package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
)

func rpcCall[T interface{}](url string, method string, args ...interface{}) (*T, error) {
	client, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	var result *T
	err = client.CallContext(context.Background(), &result, method, args...)
	if err != nil {
		return nil, err
	}
	return result, err
}
