package api

import (
	"fmt"
	"github.com/pronist/upbit/client"
	"strconv"
	"time"
)

func (api *API) GetOrderChance(market string) (map[string]interface{}, error) {
	chance, err := api.Client.CallWith("GET", "/orders/chance", client.Query{"market": market})
	if err != nil {
		return nil, err
	}

	return chance.(map[string]interface{}), nil
}

func (api *API) GetOrderList(market, state string) ([]map[string]interface{}, error) {
	orders, err := api.Client.CallWith("GET", "/orders", client.Query{
		"market": market, "state": state,
	})
	if err != nil {
		return nil, err
	}

	return client.TransformArrayMap(orders), nil
}

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

func (api *API) CancelOrder(uuid string) error {
	_, err := api.Client.CallWith("DELETE", "/order", client.Query{"uuid": uuid})
	if err != nil {
		return err
	}

	return nil
}

/////

func (api *API) Wait(done chan int, uuid string) {
	for {
		order, _ := api.Client.CallWith("GET", "/order", client.Query{"uuid": uuid})

		if order, ok := order.(map[string]interface{}); ok {
			if state, ok := order["state"].(string); ok && state == "done" {
				done <- 1
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (api *API) GetAskOrders(orders []map[string]interface{}) []map[string]interface{} {
	var o []map[string]interface{}

	for _, order := range orders {
		if order["ord_type"] == "limit" && order["side"] == "ask" {
			o = append(o, order)
		}
	}

	return o
}

func (api *API) GetLatestAskPrice(orders []map[string]interface{}) (float64, error) {
	var latestAskPrice float64
	var err error

	for _, order := range orders {
		if order["ord_type"] == "limit" && order["side"] == "ask" {
			if price, ok := order["price"].(string); ok {
				latestAskPrice, err = strconv.ParseFloat(price, 64)
				if err != nil {
					return 0, err
				}

				break
			}
		}
	}

	return latestAskPrice, nil
}
