package bot

import (
	"math"
	"time"

	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
)

// 트래킹할 종목에 대한 조건이다.
func predicate(b *Bot, t map[string]interface{}) bool {
	price := t["trade_price"].(float64)
	c := t["code"].(string)

	// https://wikidocs.net/21888
	candles, err := b.QuotationClient.Call(
		"/candles/days",
		struct {
			Market string `url:"market"`
			Count  int    `url:"count"`
		}{c, 2})
	if err != nil {
		panic(err)
	}
	dayCandles := candles.([]map[string]interface{})

	// "변동성 돌파" 한 종목을 트래킹할 조건으로 설정.
	R := dayCandles[1]["high_price"].(float64) - dayCandles[1]["low_price"].(float64)

	return dayCandles[0]["opening_price"].(float64)+(R*static.Config.K) < price
}

// 변동성 돌파전략이다. 상승장에 구입한다.
type Penetration struct {
	H float64 // 판매 상승 기준
	L float64 // 구입 하락 기준
	K float64 // 돌파 상수
}

func (p *Penetration) prepare(bot *Bot) {
	//
	log.Logger <- log.Log{Msg: "Prepare strategy...", Fields: logrus.Fields{"strategy": "Penetration"}, Level: logrus.DebugLevel}
	//

	markets, err := getMarketNames(bot, targetMarket)
	if err != nil {
		panic(err)
	}

	wsc, err := client.NewWebsocketClient("ticker", markets, true, false)
	if err != nil {
		panic(err)
	}

	if err := wsc.Ws.WriteJSON(wsc.Data); err != nil {
		panic(err)
	}

	for _, market := range markets {
		var r map[string]interface{}

		if err := wsc.Ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
		if _, ok := stat[market]; !ok && predicate(bot, r) {
			// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
			stat[market] = untracked
			//
			log.Logger <- log.Log{
				Msg: "Untracked",
				Fields: logrus.Fields{
					"market":      market,
					"change-rate": r["signed_change_rate"].(float64),
					"price":       r["trade_price"].(float64),
				},
				Level: logrus.InfoLevel,
			}
			//
		}

		time.Sleep(time.Millisecond * 300)
	}
	if err := wsc.Ws.Close(); err != nil {
		panic(err)
	}
}

func (p *Penetration) run(bot *Bot, c *coin, t map[string]interface{}) (bool, error) {
	m := t["code"].(string)
	price := t["trade_price"].(float64)

	volume := c.onceOrderPrice / price

	if math.IsInf(volume, 0) {
		panic("division by zero")
	}

	acc, err := bot.Accounts.accounts()
	if err != nil {
		return false, err
	}

	balances := getBalances(acc)

	// 변동성 돌파는 전략의 기본 조건이다.
	if predicate(bot, t) {
		if coinBalance, ok := balances[c.name]; ok {
			// 이미 코인을 가지고 있는 경우

			avgBuyPrice := getAverageBuyPrice(acc, c.name)

			pp := price / avgBuyPrice

			////// 매수 전략

			// 분할 매수 전략 (하락시 평균단가를 낮추는 전략)
			// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우

			if pp-1 <= p.L {
				return bot.Accounts.order(bot, c, b, volume, price)
			}

			////// 매도 전략

			// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우
			if pp-1 >= p.H {
				ok, err := bot.Accounts.order(bot, c, s, coinBalance, price)

				if ok && err == nil {
					// 매도 이후에는 추적 상태를 멈춘다.
					stat[m] = untracked
					//
					log.Logger <- log.Log{Msg: "Untracked", Fields: logrus.Fields{"market": c}, Level: logrus.InfoLevel}
					//
				}
				return ok, err
			}
		} else {
			// 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
			return bot.Accounts.order(bot, c, b, volume, price)
		}
	}

	return false, nil
}
