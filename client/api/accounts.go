package api

import (
	"strconv"
)

func (api *API) NewAccounts() ([]map[string]interface{}, error) {
	var acc []map[string]interface{}

	accounts, err := api.Client.Call("GET", "/accounts")
	if err != nil {
		return nil, err
	}
	if accounts, ok := accounts.([]interface{}); ok {
		for _, account := range accounts {
			if account, ok := account.(map[string]interface{}); ok {
				acc = append(acc, account)
			}
		}
	}

	return acc, nil
}

func (api *API) GetTotalBalance(accounts []map[string]interface{}, balances map[string]float64) (float64, error) {
	totalBalance := balances["KRW"]

	for coin, balance := range balances {
		if coin != "KRW" {
			avgBuyPrice, err := api.GetAverageBuyPrice(accounts, coin)
			if err != nil {
				return 0, nil
			}
			totalBalance += avgBuyPrice * balance
		}
	}

	return totalBalance, nil
}

func (api *API) GetBalances(accounts []map[string]interface{}, ) (map[string]float64, error) {
	balances := make(map[string]float64)

	for _, acc := range accounts {
		balance, err := strconv.ParseFloat(acc["balance"].(string), 64)
		if err != nil {
			return nil, err
		}

		balances[acc["currency"].(string)] = balance
	}

	return balances, nil
}

func (api *API) GetAverageBuyPrice(accounts []map[string]interface{}, coin string) (float64, error) {
	var avgBuyPrice float64
	var err error

	for _, account := range accounts {
		if currency, ok := account["currency"].(string); ok && currency == coin {
			avgBuyPrice, err = strconv.ParseFloat(account["avg_buy_price"].(string), 64)
			if err != nil {
				return 0, err
			}
			break
		}
	}

	return avgBuyPrice, nil
}
