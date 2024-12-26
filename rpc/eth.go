package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
)

type EthClient struct {
	e *ethclient.Client
	r *rpc.Client
}

func Connect(rpcUrl string) (*EthClient, error) {
	r, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	e := ethclient.NewClient(r)

	return &EthClient{e, r}, nil
}

func (client *EthClient) GetBlockNumber() (uint64, error) {
	return client.e.BlockNumber(context.TODO())
}

func (client *EthClient) GetBlockByNumber(number *big.Int) (*types.Block, error) {
	return client.e.BlockByNumber(context.TODO(), number)
}

func (client *EthClient) StaticCall(to *common.Address, data []byte, gasLimit uint64, from common.Address) ([]byte, error) {
	var gasPrice, getGasPriceErr = client.e.SuggestGasPrice(context.Background())
	if getGasPriceErr != nil {
		log.Error().Err(getGasPriceErr)
		return nil, getGasPriceErr
	}
	var msgData = ethereum.CallMsg{
		From:     from,
		To:       to,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}
	return client.e.CallContract(context.Background(), msgData, nil)
}
