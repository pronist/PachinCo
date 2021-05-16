package upbit

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var c *Client

func init() {
	c = &Client{http.DefaultClient, Config.AccessKey, Config.SecretKey}
}

func TestClient_Call(t *testing.T) {
	chance, err := c.Call("GET", "/orders/chance", struct {
		Market string `url:"market"`
	}{"KRW-BTC"})
	assert.NoError(t, err)

	assert.Contains(t, chance, "bid_fee")
	assert.Contains(t, chance, "ask_fee")
	assert.Contains(t, chance, "market")
	assert.Contains(t, chance, "bid_account")
	assert.Contains(t, chance, "ask_account")
}
