package bot

import (
	"net/http"
	"testing"

	"github.com/pronist/upbit/client"
	"github.com/stretchr/testify/assert"
)

func TestGetMarketNames(t *testing.T) {
	bot := &Bot{QuotationClient: &client.QuotationClient{Client: http.DefaultClient}}

	markets, err := getMarketNames(bot, targetMarket)
	assert.NoError(t, err)

	assert.Contains(t, markets, "KRW-BTC")
	assert.Contains(t, markets, "KRW-ETH")
	assert.NotContains(t, markets, "USDT-BTC")
	assert.NotContains(t, markets, "BTC-MKR")
}
