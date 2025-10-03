package helpers

import "math/big"

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func BigIsZero(x *big.Int) bool {
	return x == nil || x.Sign() == 0
}
