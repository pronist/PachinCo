package upbit

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"testing"
)

var data = struct{}{}

func TestSerialize(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(buf)
	assert.NoError(t, enc.Encode(data))

	r, err := Serialize(data)
	assert.NoError(t, err)

	assert.Equal(t, buf.Bytes(), r)

	_, err = Serialize(nil)
	assert.Error(t, err)
}

func TestDeserialize(t *testing.T) {
	var decodedData, r struct{}
	buf := new(bytes.Buffer)

	encodedData, err := Serialize(data)
	assert.NoError(t, err)

	buf.Write(encodedData)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(&decodedData)
	assert.NoError(t, err)

	err = Deserialize(encodedData, &r)
	assert.NoError(t, err)

	assert.Equal(t, r, decodedData)

	err = Deserialize(encodedData, r)
	assert.Error(t, err)
}
