package rpc

import (
	"encoding/json"
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
			client, err := Connect(RPC_URL)
			if err != nil {
				t.Fatal(err)
			}
			_, err = rpcCall[json.RawMessage](client, tt.method, tt.args...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
