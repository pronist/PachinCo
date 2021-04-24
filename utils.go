package upbit

import (
	"encoding/binary"
	"math"
)

func Float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)

	return float
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}