package helpers

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"strings"

	"github.com/mr-tron/base58"
)

func Encode(m any) ([]byte, error) {
	buf := bytes.Buffer{}

	err := gob.NewEncoder(&buf).Encode(m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decode(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	return gob.NewDecoder(buf).Decode(target)
}

func DecodeAnyBase(data string) ([]byte, error) {
	if strings.ContainsAny(data, "+/=") {
		return base64.StdEncoding.DecodeString(data)
	}
	// intenta base58 primero
	if b, err := base58.Decode(data); err == nil {
		return b, nil
	}
	// fallback: intenta base64
	if b, err := base64.StdEncoding.DecodeString(data); err == nil {
		return b, nil
	}
	return nil, errors.New("unknown encoding for instruction data")
}
