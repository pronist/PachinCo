package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebsocketClient(t *testing.T) {
	wsc, err := NewWebsocketClient("ticker", []string{"KRW-BTC"}, true, false)
	assert.NoError(t, err)

	assert.IsType(t, &WebsocketClient{}, wsc)
}
