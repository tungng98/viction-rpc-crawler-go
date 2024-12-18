package ethutil

import (
	"encoding/hex"
	"math/big"
	"strings"
)

func BigIntToHex(i *big.Int) string {
	return "0x" + i.Text(16)
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
