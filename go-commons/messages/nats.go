package messages

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func canonicalizeJSON(in []byte) []byte {
	var v any
	if json.Unmarshal(in, &v) != nil {
		return in
	}
	out, err := json.Marshal(v)
	if err != nil {
		return in
	}
	return out
}

func BuildProgramMsgID(slot uint64, pubkey string, data []byte) string {
	norm := canonicalizeJSON(data)
	sum := sha256.Sum256(norm)
	return fmt.Sprintf("solprog:%d:%s:%x", slot, pubkey, sum[:16])
}
