package api

import (
	"github.com/pronist/upbit/client"
)

func (api API) GetTicker(market string) ([]map[string]interface{}, error) {
	ticker, err := api.QuotationClient.Do("/ticker", client.Query{"markets": market})
	if err != nil {
		return nil, err
	}

	return client.TransformArrayMap(ticker), nil
}
