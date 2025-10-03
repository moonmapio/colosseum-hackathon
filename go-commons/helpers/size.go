package helpers

import (
	"math/big"
	"unsafe"
)

func ApproxStringSize(s string) int64 {
	return int64(unsafe.Sizeof(s)) + int64(len(s))
}

func ApproxBigInt(b *big.Int) int64 {
	if b == nil {
		return 0
	}
	bytes := int64((b.BitLen() + 7) / 8)
	return int64(unsafe.Sizeof(*b)) + bytes
}

func ApproxMapStringBigInt(m map[string]*big.Int) (count int, total int64) {
	total = int64(unsafe.Sizeof(m))
	for k, v := range m {
		count++
		total += ApproxStringSize(k) + ApproxBigInt(v)
	}
	return
}

func ApproxMapStringString(m map[string]string) (count int, total int64) {
	total = int64(unsafe.Sizeof(m))
	for k, v := range m {
		count++
		total += ApproxStringSize(k) + ApproxStringSize(v)
	}
	return
}
