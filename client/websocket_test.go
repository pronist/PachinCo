package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebsocketClient(t *testing.T) {
	_, err := NewWebsocketClient("ticker", []string{"KRW-BTC"}, true, false)
	assert.NoError(t, err)
}
