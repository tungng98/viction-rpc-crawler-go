package rpc

import (
	"testing"
)

const RPC_URL = "https://rpc.viction.xyz"

func TestRcpCallString(t *testing.T) {
	tests := []struct {
		name   string
		method string
		args   []interface{}
	}{
		{"eth_coinbase", "eth_coinbase", []interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := rpcCall[string](RPC_URL, tt.method, tt.args...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
