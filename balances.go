package upbit

import (
	"bytes"
	"encoding/gob"
)

type Balances map[string]float64

func (b *Balances) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(buf)

	err := enc.Encode(b)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeBalances(encodedBalances []byte) (Balances, error) {
	var balances Balances
	buf := new(bytes.Buffer)

	buf.Write(encodedBalances)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}
