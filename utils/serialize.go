package utils

import (
	"bytes"
	"encoding/gob"
)

// 데이터를 직렬화한다.
func Serialize(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 데이터를 역직렬화한다. received 는 반드시 포인터로 넘겨져야 한다.
func Deserialize(encodedData []byte, received interface{}) error {
	buf := new(bytes.Buffer)

	buf.Write(encodedData)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(received); err != nil {
		return err
	}

	return nil
}
