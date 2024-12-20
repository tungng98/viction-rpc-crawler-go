package ethutil

import (
	"encoding/hex"
	"math/big"
	"strings"
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
	bytes, err := hex.DecodeString(ss)
	if err != nil {
		panic(err)
	}
	return bytes
}
