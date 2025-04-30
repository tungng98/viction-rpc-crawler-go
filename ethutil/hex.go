package ethutil

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func BigIntToHex(i *big.Int) string {
	return "0x" + i.Text(16)
}

func BytesEqual(x, y []byte) bool {
	if x == nil && y == nil {
		return true
	}
	if x != nil || y != nil {
		return false
	}
	if len(x) != len(y) {
		return false
	}
	for i, _ := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func HexToBytes(s string) []byte {
	ss := s
	if strings.HasPrefix(s, "0x") {
		ss = strings.TrimPrefix(s, "0x")
	}
	if len(ss)%2 == 1 {
		ss = "0" + ss
	}
	bytes, err := hex.DecodeString(ss)
	if err != nil {
		panic(err)
	}
	return bytes
}

func PubkeyToAddress(pubkey []byte) []byte {
	addr := make([]byte, 20)
	copy(addr[:], crypto.Keccak256(pubkey[1:])[12:])
	return addr
}
