package rpc

import (
	"encoding/json"
	"math/big"
	"testing"
)

const RPC_URL = "https://rpc.viction.xyz"

func TestGetBlock(t *testing.T) {
	tests := []struct {
		number *big.Int
	}{
		{big.NewInt(71717171)},
	}
	for _, tt := range tests {
		t.Run(tt.number.String(), func(t *testing.T) {
			client, err := Connect(RPC_URL)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.GetBlockByNumber2(tt.number)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

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

func TestTraceTransaction(t *testing.T) {
	tests := []struct {
		hash string
	}{
		{"0xf71fb3d6aaa44a631f3b8af92571368f80b7e211a596ac70ddfbc8de197da38e"},
	}
	for _, tt := range tests {
		t.Run(tt.hash, func(t *testing.T) {
			client, err := Connect(FULL_ARCHIVE_NODE)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.TraceTransaction(tt.hash)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTraceBlock(t *testing.T) {
	tests := []struct {
		number *big.Int
	}{
		{big.NewInt(73736810)},
	}
	for _, tt := range tests {
		t.Run(tt.number.String(), func(t *testing.T) {
			client, err := Connect(FULL_ARCHIVE_NODE)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.TraceBlockByNumber(tt.number)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

const FULL_NODE = "http://fullnode.local:10545"
const FULL_ARCHIVE_NODE = "http://parchivenode.local:10545"
const PARTIAL_ARCHIVE_NODE = "http://farchivenode.local:10545"

func TestArchiveNode(t *testing.T) {
	tests := []struct {
		txHash      string
		block       *big.Int
		fullFull    bool
		partArchive bool
		fullArchive bool
	}{
		{"0x344e80482d2783cb5e2d319a11be13524045d89cbe42b27d525858cee1980c18", big.NewInt(80_000_000), false, false, true},
		{"0x166f57034c99337991e7731260e78da6502ef0de541fa4b8b3db1dcdb75e8af6", big.NewInt(81_000_001), false, false, true},
		{"0x6e96465db7aa8320e6df5f79f8f152f27e6bbe6f11cc277e101e0200f673dac3", big.NewInt(82_000_000), false, false, true},
		{"0xcc28831a4eba68acb546cb1985d3472d8ca72dfeea3f6f2692e2e87de3e7fd71", big.NewInt(83_000_000), false, true, true},
		{"0xcaacd461e261924866079a37129050b6338873d2659175ac7c3b68d2dfb7c10c", big.NewInt(84_000_000), false, true, true},
		{"0xdaf176a2a52f9589e0ad7e1e853682b687aeb47fab964b992bf8c24237398468", big.NewInt(85_000_000), false, true, true},
	}
	clientF, _ := Connect(FULL_NODE)
	clientPA, _ := Connect(PARTIAL_ARCHIVE_NODE)
	clientFA, _ := Connect(FULL_ARCHIVE_NODE)
	for _, tt := range tests {
		t.Run(tt.block.Text(10), func(t *testing.T) {
			_, errF := clientF.TraceTransaction(tt.txHash)
			if tt.fullFull != (errF == nil) {
				t.Fatalf("Full Node traceTransaction should be %v", tt.fullFull)
			}
			_, errPA := clientPA.TraceTransaction(tt.txHash)
			if tt.partArchive != (errPA == nil) {
				t.Fatalf("Partial Archive Node traceTransaction should be %v", tt.partArchive)
			}
			_, errFA := clientFA.TraceTransaction(tt.txHash)
			if tt.fullArchive != (errFA == nil) {
				t.Fatalf("Full Archive Node traceTransaction should be %v", tt.fullFull)
			}
		})
	}
}
