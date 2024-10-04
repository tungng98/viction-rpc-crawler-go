package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

func StaticCall(client *ethclient.Client, to *common.Address, data []byte, gasLimit uint64, from common.Address) ([]byte, error) {
	var gasPrice, getGasPriceErr = client.SuggestGasPrice(context.Background())
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
	return client.CallContract(context.Background(), msgData, nil)
}
