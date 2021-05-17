package upbit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var acc = []map[string]interface{}{
	{
		"currency":               "KRW",
		"balance":                "55000.0",
		"locked":                 "0.0",
		"avg_buy_price":          "0",
		"avg_buy_price_modified": false,
		"unit_currency":          "KRW",
	},
	{
		"currency":               "BTC",
		"balance":                "2.0",
		"locked":                 "0.0",
		"avg_buy_price":          "101000",
		"avg_buy_price_modified": false,
		"unit_currency":          "KRW",
	},
}

func TestGetBalances(t *testing.T) {
	balances := getBalances(acc)

	assert.Equal(t, 55000.0, balances["KRW"])
	assert.Equal(t, 2.0, balances["BTC"])
}

func TestGetAverageBuyPrice(t *testing.T) {
	avgBuyPrice := getAverageBuyPrice(acc, "BTC")
	assert.Equal(t, 101000.0, avgBuyPrice)
}

func TestGetTotalBalance(t *testing.T) {
	avgBuyPrice := getAverageBuyPrice(acc, "BTC")
	balances := getBalances(acc)
	totalBalance := getTotalBalance(acc, balances)

	assert.Equal(t, balances["KRW"]+balances["BTC"]*avgBuyPrice, totalBalance)
}
