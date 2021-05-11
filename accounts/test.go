package accounts

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
)

// 하나의 계정을 나타낸다.
// upbit.Accounts 에 대해서는 별도로 변환하지 않기 때문에 테스트용도로만 쓴다.
type Account struct {
	Currency    string  `structs:"currency"`
	Balance     float64 `structs:"balance"`
	AvgBuyPrice float64 `structs:"avg_buy_price"`
}

type TestAccounts struct {
	accounts map[string]*Account
	krw      map[string]float64
}

func NewTestAccounts(krw float64) *TestAccounts {
	accounts := map[string]*Account{
		"KRW": {Currency: "KRW", Balance: krw, AvgBuyPrice: 0},
	}
	//
	log.Logger <- log.Log{
		Msg: "Creating new accounts for Testing.",
		Fields: logrus.Fields{
			"KRW": krw,
		},
		Level: logrus.DebugLevel,
	}
	//

	return &TestAccounts{accounts, make(map[string]float64)}
}

func (acc *TestAccounts) Order(coin *bot.Coin, side string, volume, price float64, t map[string]interface{}) (bool, error) {
	if _, ok := acc.accounts[coin.Name]; !ok {
		acc.accounts[coin.Name] = &Account{Currency: coin.Name, Balance: 0, AvgBuyPrice: 0}
	}

	var sign float64

	sign = 1
	if side == upbit.S {
		sign = -sign
	}

	acc.accounts[coin.Name].Balance += sign * volume
	acc.krw[coin.Name] += sign * volume * price
	acc.accounts[coin.Name].AvgBuyPrice = acc.krw[coin.Name] / acc.accounts[coin.Name].Balance

	acc.accounts["KRW"].Balance += -sign * volume * price

	//
	log.Logger <- log.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"coin": coin.Name, "side": side, "volume": volume, "price": price, "change-rate": t["change_rate"].(float64),
		},
		Level: logrus.WarnLevel,
	}
	//

	err := coin.Refresh(acc)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (acc *TestAccounts) Accounts() []map[string]interface{} {
	r := make([]map[string]interface{}, 0)

	for _, account := range acc.accounts {
		m := structs.Map(account)

		for k, v := range m {
			if v, ok := v.(float64); ok {
				m[k] = fmt.Sprintf("%f", v)
			}
		}
		r = append(r, m)
	}

	return r
}
