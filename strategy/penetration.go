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
func (p *Penetration) Prepare() {
	//
	log.Logger <- log.Log{Msg: "Preparing...", Fields: logrus.Fields{"strategy": "Penetration"}, Level: logrus.InfoLevel}
	//
	ws, _, err := websocket.DefaultDialer.Dial(bot.SockURL+"/"+bot.SockVersion, nil)
	if err != nil {
		panic(err)
	}

	markets, err := upbit.API.GetMarketNames(bot.TargetMarket)
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
		if _, ok := upbit.MarketTrackingStates[market]; !ok && bot.Predicate(market, r) {
			// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
			upbit.MarketTrackingStates[market] = upbit.STOPPED

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

func (p *Penetration) Run(coin *bot.Coin) {
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
	stat, ok := upbit.MarketTrackingStates[bot.TargetMarket+"-"+coin.Name]

	for ok && stat == upbit.TRACKING {
		ticker := <-coin.T

		m := ticker["code"].(string)
		price := ticker["trade_price"].(float64)

		balances, err := upbit.API.GetBalances(upbit.Accounts)
		if err != nil {
			panic(err)
		}

		if balances["KRW"] >= upbit.MinimumOrderPrice && balances["KRW"] > coin.OnceOrderPrice && coin.OnceOrderPrice > upbit.MinimumOrderPrice {
			volume := coin.OnceOrderPrice / price

			if math.IsInf(volume, 0) {
				panic(err)
			}

			// 변동성 돌파는 전략의 기본 조건이다.
			if bot.Predicate(m, ticker) {
				//if coinBalance, ok := balances[coin.Name]; ok {
				// 테스트
				if coinBalance, ok := upbit.T_Accounts[coin.Name]; ok {
					// 이미 코인을 가지고 있는 경우

					//avgBuyPrice, err := upbit.API.GetAverageBuyPrice(upbit.Accounts, coin.Name)
					//if err != nil {
					//	panic(err)
					//}

					// 테스트
					var avgBuyPrice float64
					var count float64

					for c, order := range upbit.T_Orders {
						if order.Side == upbit.B && c == coin.Name {
							count += order.Volume
							avgBuyPrice += order.Price * order.Volume
						}
					}
					avgBuyPrice /= count
					//

					pp := price / avgBuyPrice

					// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우, 단기 시장이 하락장일 때 '매도'
					if pp-1 >= p.H {
						if ok := coin.Order(upbit.S, coinBalance, price); ok {
							// 매도 이후에는 추적 상태를 멈춘다.
							upbit.MarketTrackingStates[m] = upbit.STOPPED
						}
						continue
					}
					// 가격이 매수평균가 대비 `p.L` 보다 하락한 경우 추가 '매수' 요청
					if pp-1 <= p.L {
						coin.Order(upbit.B, volume, price)
						continue
					}
				} else {
					// 단기 상승장 + 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
					coin.Order(upbit.B, volume, price)
					continue
				}
			}
		}

		time.Sleep(time.Second * 1)
	}

	log.Logger <- log.Log{
		Msg: "Strategy closed.",
		Fields: logrus.Fields{
			"strategy": "Penetration", "coin": coin.Name,
		},
		Level: logrus.WarnLevel,
	}
}
