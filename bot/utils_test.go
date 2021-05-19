package bot

import (
	"net/http"
	"testing"

	"github.com/pronist/upbit/client"
	"github.com/stretchr/testify/assert"
)

func TestGetMarketNames(t *testing.T) {
	bot := &Bot{QuotationClient: &client.QuotationClient{Client: http.DefaultClient}}

	markets, err := bot.QuotationClient.Call("/market/all", struct {
		IsDetail bool `url:"isDetail"`
	}{false})
	if err != nil {
		panic(err)
	}

	targetMarkets := getMarketNames(markets.([]map[string]interface{}), targetMarket)

	assert.Contains(t, targetMarkets, "KRW-BTC")
	assert.Contains(t, targetMarkets, "KRW-ETH")
	assert.NotContains(t, targetMarkets, "USDT-BTC")
	assert.NotContains(t, targetMarkets, "BTC-MKR")
}
