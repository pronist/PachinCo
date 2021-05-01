package api

import (
	"fmt"
	"github.com/pronist/upbit/client"
	"time"
)

func (api *API) Order(market, side string, volume, price float64) (string, error) {
	q := client.Query{
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
		if uuid, ok := order["uuid"].(string); ok {
			return uuid, nil
		}
	}

	return "", fmt.Errorf("%#v", order)
}

func (api *API) WaitUntilCompletedOrder(uuid string) error {
	for {
		order, err := api.Client.CallWith("GET", "/order", client.Query{"uuid": uuid})
		if err != nil {
			return err
		}

		if order, ok := order.(map[string]interface{}); ok {
			if state, ok := order["state"].(string); ok && state == "done" {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}
}