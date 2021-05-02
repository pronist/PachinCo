package api

import (
	"fmt"
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

/////

func (api *API) GetChangeRate(market string) (float64, error) {
	days, err := api.QuotationClient.Do("/candles/days", map[string]string{
		"market": market, "count": "1",
	})
	if err != nil {
		return 0, err
	}

	return client.GetValueFromArray(days, 0, "change_rate").(float64), fmt.Errorf("%#v", days)
}
