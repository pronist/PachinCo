package accounts

import (
	"fmt"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
	"strconv"
)

type TestAccounts struct {
	accounts []map[string]interface{}
	orders []map[string]interface{}
}

func NewTestAccounts(krw string) *TestAccounts {
	accounts := []map[string]interface{}{
		{
			"currency": "KRW",
			"balance": krw, // 500만
			"avg_buy_price": "0",
		},
	}
	//
	log.Logger <- log.Log{Msg: "Set Accounts for Testing.", Level: logrus.InfoLevel}
	//

	return &TestAccounts{accounts, make([]map[string]interface{}, 0)}
}

func (acc *TestAccounts) Order(coin *bot.Coin, side string, volume, price float64) (bool, error) {
	c := upbit.Market + "-" + coin.Name

	log.Logger <- log.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": c, "volume": volume, "price": price,
		},
		Level: logrus.WarnLevel,
	}

	acc.orders = append(acc.orders, map[string]interface{}{
		"market": c, "side": side, "price": price, "volume": volume,
	})

	for _, account := range acc.accounts {
		if account["currency"] == coin.Name {
			f, err := strconv.ParseFloat(account["balance"].(string), 64)
			if err != nil {
				log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
			}

			switch side {
			case upbit.B:
				f += volume
				account["balance"] = fmt.Sprintf("%f", f)
				//account["avg_buy_price"] = fmt.Sprintf("%f", getAverageBuyPrice(Orders, coinName))
			case upbit.S:
				f -= volume
				account["balance"] = fmt.Sprintf("%f", f)
			}

			break
		}
	}

	return true, nil
}

func (acc *TestAccounts) Accounts() []map[string]interface{} {
	return acc.accounts
}

//func getAverageBuyPrice(orders map[string]Order, coinName string) float64 {
//	var avgBuyPrice float64
//	var count float64
//
//	for c, order := range orders {
//		if order.Side == upbit.B && c == coinName {
//			count += order.Volume
//			avgBuyPrice += order.Price * order.Volume
//		}
//	}
//	avgBuyPrice /= count
//
//	return avgBuyPrice
//}