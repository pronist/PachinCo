package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWebsocketClient(t *testing.T) {
	wsc, err := NewWebsocketClient("ticker", []string{"KRW-BTC"}, true, false)
	assert.NoError(t, err)

	assert.IsType(t, &WebsocketClient{}, wsc)
}
