package upbit

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"math"
	"strings"
	"time"
)

// 트래킹할 종목에 대한 조건이다.
func predicate(b *Bot, t map[string]interface{}) bool {
	price := t["trade_price"].(float64)
	c := t["code"].(string)

	// https://wikidocs.net/21888
	candles, err := b.QuotationClient.call(
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

	return dayCandles[0]["opening_price"].(float64)+(R*Config.K) < price
}

// 변동성 돌파전략이다. 상승장에 구입한다.
type Penetration struct {
	H float64 // 판매 상승 기준
	L float64 // 구입 하락 기준
	K float64 // 돌파 상수
}

func (p *Penetration) prepare(bot *Bot, _ Accounts) {
	//
	logger <- log{msg: "Prepare strategy...", fields: logrus.Fields{"strategy": "Penetration"}, level: logrus.DebugLevel}
	//
	ws, _, err := websocket.DefaultDialer.Dial(sockURL+"/"+sockVersion, nil)
	if err != nil {
		panic(err)
	}

	markets, err := bot.QuotationClient.call("/market/all", struct{ isDetail bool }{false})
	if err != nil {
		panic(err)
	}

	targetmarkets := funk.Chain(markets.([]map[string]interface{})).
		Map(func(market map[string]interface{}) string { return market["market"].(string) }).
		Filter(func(market string) bool { return strings.HasPrefix(market, market) }).
		Value().([]string)

	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": targetmarkets, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	if err := ws.WriteJSON(data); err != nil {
		panic(err)
	}

	for _, market := range targetmarkets {
		var r map[string]interface{}

		if err := ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
		if _, ok := marketTrackingStates[market]; !ok && predicate(bot, r) {
			// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
			marketTrackingStates[market] = stopped

			//
			logger <- log{
				msg: "Stopped",
				fields: logrus.Fields{
					"market":      market,
					"change-rate": r["signed_change_rate"].(float64),
					"price":       r["trade_price"].(float64),
				},
				level: logrus.InfoLevel,
			}
			//
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (p *Penetration) run(bot *Bot, accounts Accounts, c *coin, t map[string]interface{}) (bool, error) {
	price := t["trade_price"].(float64)
	m := t["code"].(string)

	volume := c.onceOrderPrice / price

	if math.IsInf(volume, 0) {
		panic("division by zero")
	}

	acc, err := accounts.accounts()
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
				return accounts.order(bot, c, b, volume, price)
			}

			////// 매도 전략

			// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우
			if pp-1 >= p.H {
				ok, err := accounts.order(bot, c, s, coinBalance, price)

				if ok && err == nil {
					// 매도 이후에는 추적 상태를 멈춘다.
					marketTrackingStates[m] = stopped
					//
					logger <- log{
						msg:    "Stopped",
						fields: logrus.Fields{"market": c},
						level:  logrus.InfoLevel,
					}
					//
				}
				return ok, err
			}
		} else {
			// 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
			return accounts.order(bot, c, b, volume, price)
		}
	}

	return false, nil
}
