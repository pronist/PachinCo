package bot

import (
	"math"
	"time"

	"github.com/pronist/upbit/client"

	"github.com/jasonlvhit/gocron"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
)

const K = 0.5 // 돌파 상수

// 트래킹할 종목에 대한 조건이다.
func predicate(b *Bot, t map[string]interface{}) bool {
	price := t["trade_price"].(float64)
	c := t["code"].(string)

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

	return dayCandles[0]["opening_price"].(float64)+(R*K) < price
}

// 변동성 돌파전략이다. 상승장에 구입한다.
type PenetrationStrategy struct{}

// 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
func (p *PenetrationStrategy) register(bot *Bot) error {
	markets, err := bot.QuotationClient.Call("/market/all", struct {
		IsDetail bool `url:"isDetail"`
	}{false})
	if err != nil {
		panic(err)
	}

	targetMarkets, err := getMarketNames(markets.([]map[string]interface{}), targetMarket)
	if err != nil {
		panic(err)
	}

	wsc, err := client.NewWebsocketClient("ticker", targetMarkets, true, false)
	if err != nil {
		return err
	}

	if err := wsc.Ws.WriteJSON(wsc.Data); err != nil {
		return err
	}

	for _, market := range targetMarkets {
		var r map[string]interface{}

		if err := wsc.Ws.ReadJSON(&r); err != nil {
			return err
		}

		// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
		if _, ok := stat[market]; !ok && predicate(bot, r) {
			stat[market] = excluded

			log.Logger <- log.Log{
				Msg: "Excluded",
				Fields: logrus.Fields{
					"market":      market,
					"change-rate": r["signed_change_rate"].(float64),
					"price":       r["trade_price"].(float64),
				},
				Level: logrus.InfoLevel,
			}
		}

		time.Sleep(time.Millisecond * 300)
	}

	return wsc.Ws.Close()
}

// 오전 9시에 매도 주문을 낸다.
func (p *PenetrationStrategy) boot(bot *Bot, c *coin) error {
	s := gocron.NewScheduler()

	if err := s.Every(1).Day().At("09:00").Do(p.s, bot, c, s); err != nil {
		return err
	}
	s.Start()

	return nil
}

func (p *PenetrationStrategy) run(bot *Bot, c *coin, t map[string]interface{}) (bool, error) {
	price := t["trade_price"].(float64)

	volume := c.onceOrderPrice / price

	if math.IsInf(volume, 0) {
		panic("division by zero")
	}

	acc, err := bot.accounts.accounts()
	if err != nil {
		return false, err
	}

	balances := getBalances(acc)

	// 변동성 돌파는 전략의 기본 조건이다.
	if predicate(bot, t) {
		if _, ok := balances[c.name]; ok {
			// 이미 코인을 가지고 있는 경우
		} else {
			// 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
			return bot.accounts.order(bot, c, b, volume, price)
		}
	}

	return false, nil
}

// 매도 전략
func (p *PenetrationStrategy) s(bot *Bot, c *coin, scheduler *gocron.Scheduler) error {
	ticker, err := bot.QuotationClient.Call("/ticker", struct {
		Markets string `url:"markets"`
	}{targetMarket + "-" + c.name})
	if err != nil {
		return err
	}

	t := ticker.([]map[string]interface{})

	// 변동성 돌파전략에 다음날 '시가' 에 처분한다.
	openingPrice := t[0]["opening_price"].(float64)

	acc, err := bot.accounts.accounts()
	if err != nil {
		return err
	}

	balances := getBalances(acc)

	if coinBalance, ok := balances[c.name]; ok {
		_, err := bot.accounts.order(bot, c, s, coinBalance, openingPrice)

		if err == nil {
			// 매도 이후에는 추적 상태를 멈춘다.
			// 시가에 처분한 이후 다시 변동성 돌파하면 언제든 다시 추적할 수 있다.
			stat[t[0]["market"].(string)] = untracked

			log.Logger <- log.Log{
				Msg:    "Untracked",
				Fields: logrus.Fields{"market": c},
				Level:  logrus.InfoLevel,
			}

			// 매도를 하여 추적을 종료 할 것이므로 같이 날린다.
			scheduler.Remove(p.s)
		}
	}

	return nil
}
