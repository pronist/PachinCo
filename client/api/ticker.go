package api

import (
	"github.com/pronist/upbit/client"
)

/////

func (api *API) GetPrice(market string) (float64, error) {
	ticker, err := api.QuotationClient.Do("/ticker", client.Query{"markets": market})
	if err != nil {
		return 0, err
	}

	return client.GetValueFromArray(ticker, 0, "trade_price").(float64), nil
}
