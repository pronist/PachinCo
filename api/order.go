package api

import (
	"fmt"
	"github.com/pronist/upbit/gateway"
)

func (api *API) Order(market, side string, volume, price float64) (string, error) {
	q := gateway.Query{
		"market":   market,
		"side":     side,
		"volume":   fmt.Sprintf("%f", volume),
		"price":    fmt.Sprintf("%f", price),
		"ord_type": "limit",
	}
	order, err := api.Client.CallWith("POST", "/orders", q)
	if err != nil {
		return "", err
	}

	if order, ok := order.(map[string]interface{}); ok {
		return order["uuid"].(string), nil
	}

	return "", err
}
