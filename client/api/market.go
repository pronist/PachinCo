package api

import (
	"github.com/pronist/upbit/client"
)

func (api *API) GetMarkets() ([]map[string]interface{}, error) {
	var m []map[string]interface{}

	markets, err := api.QuotationClient.Do("/market/all", client.Query{
		"isDetails": "false",
	})
	if err != nil {
		return nil, err
	}

	if markets, ok := markets.([]interface{}); ok {
		for _, market := range markets {
			if market, ok := market.(map[string]interface{}); ok {
				m = append(m, market)
			}
		}
	}

	return m, nil
}
