package helpers

import (
	"errors"

	"github.com/mr-tron/base58"
)

type Pubkey [32]byte

func (p Pubkey) String() string { return base58.Encode(p[:]) }

func Base58ToPublicKey(s string) (Pubkey, error) {
	b, err := base58.Decode(s)
	if err != nil || len(b) != 32 {
		return Pubkey{}, errors.New("invalid pubkey")
	}
	var p Pubkey
	copy(p[:], b)
	return p, nil
}
