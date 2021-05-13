package utils

import (
	"bytes"
	"encoding/gob"
)

func Serialize(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Deserialize(encodedData []byte, received interface{}) error {
	buf := new(bytes.Buffer)

	buf.Write(encodedData)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(received); err != nil {
		return err
	}

	return nil
}