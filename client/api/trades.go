package api

import (
	"github.com/pronist/upbit/client"
	"strconv"
)

func (api *API) GetTrades(market string, count int) ([]map[string]interface{}, error) {
	var t []map[string]interface{}

	trades, err := api.QuotationClient.Do("/trades/ticks", client.Query{
		"market": market, "count": strconv.Itoa(count),
	})
	if err != nil {
		return nil, err
	}

	if trades, ok := trades.([]interface{}); ok {
		for _, trade := range trades {
			if trade, ok := trade.(map[string]interface{}); ok {
				t = append(t, trade)
			}
		}
	}

	return t, nil
}