package api

import (
	"github.com/pronist/upbit/client"
)

func (api *API) GetCandlesMinutes(unit, market, count string) ([]map[string]interface{}, error) {
	minutesCandles, err := api.QuotationClient.Do("/candles/minutes/"+unit, client.Query{
		"market": market, "count": count,
	})

	if err != nil {
		return nil, err
	}

	return client.TransformArrayMap(minutesCandles), nil
}

func (api *API) GetCandlesDays(market, count string) ([]map[string]interface{}, error) {
	daysCandles, err := api.QuotationClient.Do("/candles/days", map[string]string{
		"market": market, "count": count,
	})
	if err != nil {
		return nil, err
	}

	return client.TransformArrayMap(daysCandles), nil
}
