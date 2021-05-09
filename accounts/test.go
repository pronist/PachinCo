package accounts

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
)

type TestAccounts struct {
	accounts []map[string]interface{}
	orders   []map[string]interface{}
}

func NewTestAccounts(krw string) *TestAccounts {
	accounts := []map[string]interface{}{
		{
			"currency":      "KRW",
			"balance":       krw, // 500만
			"avg_buy_price": "0",
		},
	}
	//
	log.Logger <- log.Log{
		Msg: "Creating new accounts for Testing.",
		Fields: logrus.Fields{
			"KRW": krw,
		},
		Level: logrus.InfoLevel,
	}
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

	//for range acc.accounts {}

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
