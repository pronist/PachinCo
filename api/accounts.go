package api

import (
	"strconv"
)

type Accounts struct {
	accounts []map[string]interface{}
}

func (api *API) NewAccounts() (*Accounts, error) {
	acc := Accounts{}

	accounts, err := api.Client.Call("GET", "/accounts")
	if err != nil {
		return nil, err
	}
	if accounts, ok := accounts.([]interface{}); ok {
		for _, account := range accounts {
			if account, ok := account.(map[string]interface{}); ok {
				acc.accounts = append(acc.accounts, account)
			}
		}
	}

	return &acc, nil
}

func (acc *Accounts) GetTotalBalance(balances map[string]float64) (float64, error) {
	totalBalance := balances["KRW"]

	delete(balances, "KRW")

	for coin, balance := range balances {
		avgBuyPrice, err := acc.GetAverageBuyPrice(coin)
		if err != nil {
			return 0, nil
		}
		totalBalance += avgBuyPrice * balance
	}

	return totalBalance, nil
}

func (acc *Accounts) GetBalances() (map[string]float64, error) {
	balances := make(map[string]float64)

	for _, acc := range acc.accounts {
		balance, err := strconv.ParseFloat(acc["balance"].(string), 64)
		if err != nil {
			return nil, err
		}

		balances[acc["currency"].(string)] = balance
	}

	return balances, nil
}

func (acc *Accounts) GetAverageBuyPrice(coin string) (float64, error) {
	var avgBuyPrice float64

	for _, account := range acc.accounts {
		if account["currency"].(string) == coin {
			var err error

			avgBuyPrice, err = strconv.ParseFloat(account["avg_buy_price"].(string), 64)
			if err != nil {
				return 0, err
			}
			break
		}
	}

	return avgBuyPrice, nil
}
