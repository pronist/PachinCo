package api

import (
	"github.com/pronist/upbit/client"
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

/////