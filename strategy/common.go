package strategy

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/sirupsen/logrus"
	"time"
)

// update 메서드는 봇과 업비트와의 계좌 동기화를 위해 정보를 갱신해야 한다.
// 주로 매수/매도를 할 때 정보의 변동이 발생하므로 주문 이후 즉시 처리한다.
func update(coin string, coinRate float64) ([]map[string]interface{}, map[string]float64, float64, float64) {
	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := upbit.API.NewAccounts()
	if err != nil {
		bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
	}
	balances, err := upbit.API.GetBalances(accounts)
	if err != nil {
		bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
	}

	// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
	totalBalance, err := upbit.API.GetTotalBalance(accounts, balances) // 초기 자금
	if err != nil {
		bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
	}

	// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
	// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
	// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
	limitOrderPrice := totalBalance * coinRate

	// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
	// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
	orderBuyingPrice := limitOrderPrice * upbit.Config.R

	bot.LogChan <- bot.Log{
		Msg: coin,
		Fields: logrus.Fields{
			"total-balance": totalBalance, "limit-order-price": limitOrderPrice, "order-buying-price": orderBuyingPrice,
		},
		Level: logrus.InfoLevel,
	}

	return accounts, balances, limitOrderPrice, orderBuyingPrice
}

// order 메서드는 주문을 하되 Config.Timeout 만큼이 지나가면 주문을 자동으로 취소한다.
// 매수/매도에 둘다 사용한다.
func order(coin, side string, coinRate float64, volume, price float64) ([]map[string]interface{}, map[string]float64, float64, float64) {
	uuid, err := upbit.API.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
	}
	bot.LogChan <- bot.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": "KRW-" + coin, "volume": volume, "price": price,
		},
		Level: logrus.WarnLevel,
	}

	done := make(chan int)

	timer := time.NewTimer(time.Second * upbit.Config.Timeout)

	go upbit.API.Wait(done, uuid)

	select {
	// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
	case <-timer.C:
		err := upbit.API.CancelOrder(uuid)
		if err != nil {
			bot.LogChan <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
		}
		bot.LogChan <- bot.Log{
			Msg: "CANCEL",
			Fields: logrus.Fields{
				"coin": coin, "side": side, "timeout": time.Second * upbit.Config.Timeout,
			},
			Level: logrus.WarnLevel,
		}
		return update(coin, coinRate)
	case <-done:
		return update(coin, coinRate)
	}
}
