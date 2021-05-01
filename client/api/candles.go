package api

import "fmt"

// 전일 종가대비 변화율
func (api *API) GetChangeRate(market string) (float64, error) {
	day, err := api.QuotationClient.Do("/candles/days", map[string]string{
		"market": market, "count": "1",
	})
	if err != nil {
		return 0, err
	}
	if day, ok := day.([]interface{}); ok {
		if candle, ok := day[0].(map[string]interface{}); ok {
			if changeRate, ok := candle["change_rate"].(float64); ok {
				return changeRate, nil
			}
		}
	}

	return 0, fmt.Errorf("%#v", day)
}
