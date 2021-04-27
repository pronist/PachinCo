package api

import "github.com/pronist/upbit/gateway"

func (api *API) GetPrice(market string) (float64, error) {
	ticker, err := api.QuotationClient.Do("/ticker", gateway.Query{"markets": market})
	if err != nil {
		return 0, err
	}
	if ticker, ok := ticker.([]interface{}); ok {
		if tick, ok := ticker[0].(map[string]interface{}); ok {
			return tick["trade_price"].(float64), nil
		}
	}

	return 0, err
}
