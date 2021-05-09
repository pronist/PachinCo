package strategy

import (
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

// 변동성 돌파전략이다. 상승장에 구입한다.
type Penetration struct {
	H float64 // 판매 상승 기준
	L float64 // 구입 하락 기준
	K float64 // 돌파 상수
}

// 이미 돌파된 종목에 대해서는 처리하면 안 된다.
func (p *Penetration) Prepare(accounts bot.Accounts) {
	//
	log.Logger <- log.Log{Msg: "Prepare strategy...", Fields: logrus.Fields{"strategy": "Penetration"}, Level: logrus.InfoLevel}
	//
	ws, _, err := websocket.DefaultDialer.Dial(bot.SockURL+"/"+bot.SockVersion, nil)
	if err != nil {
		panic(err)
	}

	markets, err := upbit.API.GetMarketNames(upbit.Market)
	if err != nil {
		panic(err)
	}

	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": markets, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	if err := ws.WriteJSON(data); err != nil {
		panic(err)
	}

	for _, market := range markets {
		var r map[string]interface{}

		if err := ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
		if _, ok := bot.MarketTrackingStates[market]; !ok && bot.Predicate(market, r) {
			// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
			bot.MarketTrackingStates[market] = bot.STOPPED

			//
			log.Logger <- log.Log{
				Msg: "Stopped",
				Fields: logrus.Fields{
					"market":      market,
					"change-rate": r["signed_change_rate"].(float64),
					"price":       r["trade_price"].(float64),
				},
				Level: logrus.WarnLevel,
			}
			//
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (p *Penetration) Run(accounts bot.Accounts, coin *bot.Coin) {
	defer func(coin *bot.Coin) {
		if err := recover(); err != nil {
			//
			log.Logger <- log.Log{
				Msg: err,
				Fields: logrus.Fields{
					"role": "Strategy", "strategy": "Penetration", "coin": coin.Name,
				},
				Level: logrus.ErrorLevel,
			}
			//
		}
	}(coin)

	//
	log.Logger <- log.Log{
		Msg: "Strategy started...",
		Fields: logrus.Fields{
			"strategy": "Penetration", "coin": coin.Name,
		},
		Level: logrus.WarnLevel,
	}
	//
	stat, ok := bot.MarketTrackingStates[upbit.Market+"-"+coin.Name]

	for ok && stat == bot.TRACKING {
		ticker := <-coin.T

		price := ticker["trade_price"].(float64)
		c := ticker["code"].(string)

		acc := accounts.Accounts()

		balances, err := upbit.API.GetBalances(acc)
		if err != nil {
			panic(err)
		}

		if balances["KRW"] >= upbit.MinimumOrderPrice && balances["KRW"] > coin.OnceOrderPrice && coin.OnceOrderPrice > upbit.MinimumOrderPrice {
			volume := coin.OnceOrderPrice / price

			if math.IsInf(volume, 0) {
				panic(err)
			}

			// 변동성 돌파는 전략의 기본 조건이다.
			if bot.Predicate(c, ticker) {
				if coinBalance, ok := balances[coin.Name]; ok {
					// 이미 코인을 가지고 있는 경우

					avgBuyPrice, err := upbit.API.GetAverageBuyPrice(acc, coin.Name)
					if err != nil {
						panic(err)
					}

					pp := price / avgBuyPrice

					////// 매수 전략

					// 분할 매수 전략 (하락시 평균단가를 낮추는 전략)
					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우

					if avgBuyPrice*coinBalance+coin.OnceOrderPrice <= coin.Limit {
						if pp-1 <= p.L {
							accounts.Order(coin, upbit.B, volume, price)
							continue
						}
					}

					////// 매도 전략

					// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우

					orderSellingPrice := coinBalance * price

					if pp-1 >= p.H && orderSellingPrice > upbit.MinimumOrderPrice {
						if ok, err := accounts.Order(coin, upbit.S, coinBalance, price); ok && err == nil {
							// 매도 이후에는 추적 상태를 멈춘다.
							bot.MarketTrackingStates[c] = bot.STOPPED
						}
						continue
					}
				} else {
					// 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
					accounts.Order(coin, upbit.B, volume, price)
					continue
				}
			}
		}

		time.Sleep(time.Second * 1)
	}
}
