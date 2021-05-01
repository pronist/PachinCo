package api

import (
	"fmt"
	"github.com/pronist/upbit/client"
)

func (api *API) GetPrice(market string) (float64, error) {
	ticker, err := api.QuotationClient.Do("/ticker", client.Query{"markets": market})
	if err != nil {
		return 0, err
	}
	if ticker, ok := ticker.([]interface{}); ok {
		if tick, ok := ticker[0].(map[string]interface{}); ok {
			if tradePrice, ok := tick["trade_price"].(float64); ok {
				return tradePrice, nil
			}
		}
	}

	return 0, fmt.Errorf("%#v", ticker)
}
