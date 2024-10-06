package ethutil

import (
	"time"

	"golang.org/x/exp/rand"
)

func init() {
	rand.Seed(uint64(time.Now().Unix()))
}

var hexRunes = []rune("0123456789abcdef")

func RandomAddress() string {
	return RandomHex(40)
}

func RandomBlockHash() string {
	return RandomHex(64)
}

func RandomHex(length uint8) string {
	b := make([]rune, length+2)
	for i := range b {
		b[i] = hexRunes[rand.Intn(16)]
	}
	b[0] = '0'
	b[1] = 'x'
	return string(b)
}

func RandomNumber(min, max uint64) uint64 {
	return min + (rand.Uint64() % (max - min))
}

func RandomTxHash() string {
	return RandomHex(64)
}
