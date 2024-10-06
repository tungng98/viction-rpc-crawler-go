package ethutil

import "math/big"

func BigIntToHex(i *big.Int) string {
	return "0x" + i.Text(16)
}
