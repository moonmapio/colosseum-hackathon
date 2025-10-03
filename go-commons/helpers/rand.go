package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func RandomInt(min, max int64) (int64, error) {
	if max < min {
		return 0, fmt.Errorf("invalid range: max < min")
	}
	span := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return 0, err
	}
	return min + n.Int64(), nil
}

func RandomNumber() uint64 {
	const min int64 = 1
	const max int64 = 100_000
	var lastErr error
	for i := 0; i < 10; i++ {
		if v, err := RandomInt(min, max); err == nil {
			return uint64(v)
		} else {
			lastErr = err
		}
	}
	panic(fmt.Errorf("RandomNumber failed after 10 attempts: %w", lastErr))
}
