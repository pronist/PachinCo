package api

import (
	"github.com/pronist/upbit/client"
	"github.com/thoas/go-funk"
	"strings"
)

func (api *API) GetMarkets() ([]map[string]interface{}, error) {
	markets, err := api.QuotationClient.Do("/market/all", client.Query{
		"isDetails": "false",
	})
	if err != nil {
		return nil, err
	}

	return client.TransformArrayMap(markets), nil
}

func (api *API) GetMarketNames(currency string) ([]string, error) {
	markets, err := api.GetMarkets()
	if err != nil {
		return nil, err
	}

	return funk.Chain(markets).
		Map(func(market map[string]interface{}) string {
			return market["market"].(string)
		}).
		Filter(func(market string) bool {
			return strings.HasPrefix(market, currency)
		}).
		Value().([]string), nil
}
//

func (api *API) GetMarketConditionBy(candles []map[string]interface{}, price float64) bool {
	var totalClosePrice float64

	// 지금은 제외
	for _, candle := range candles[1:] {
		totalClosePrice += candle["trade_price"].(float64)
	}

	return totalClosePrice/float64(len(candles)-1) < price
}
