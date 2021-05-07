package strategy

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
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
func init() {
	bot.Logger <- bot.Log{
		Msg: "Initialization strategy...",
		Fields: logrus.Fields{
			"Strategy": "Penetration",
		},
		Level: logrus.InfoLevel,
	}

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

		err = upbit.Db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

			// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
			if bot.Predicate(market, r) && bytes.Compare(bucket.Get([]byte(market[4:])), []byte{bot.TRACKING}) != 0 {
			// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
				if err := bucket.Put([]byte(market[4:]), []byte{bot.STOPPED}); err != nil {
					panic(err)
				}

				bot.Logger <- bot.Log{
					Msg: "State change to `STOPPED`",
					Fields: logrus.Fields{
						"coin": market[4:],
					},
					Level: logrus.WarnLevel,
				}
			}

			return nil
		})
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (p *Penetration) Run(coin *bot.Coin) {
	defer func(coin *bot.Coin) {
		if err := recover(); err != nil {
			bot.Logger <- bot.Log{
				Msg: err,
				Fields: logrus.Fields{
					"role": "Strategy", "strategy": "Penetration", "coin": coin.Name,
				},
				Level: logrus.ErrorLevel,
			}
		}
	}(coin)

	bot.Logger <- bot.Log{
		Msg: "Strategy started...",
		Fields: logrus.Fields{
			"strategy": "Penetration", "coin": coin.Name,
		},
		Level: logrus.InfoLevel,
	}

	var stat []byte

	// 먼저 버킷에 있는 코인의 상태를 얻어오자.
	err := upbit.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		stat = bucket.Get([]byte(coin.Name))
		if stat == nil {
			return fmt.Errorf("Not found %#v in bucket", coin.Name)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	for bytes.Equal(stat, []byte{bot.TRACKING}) {
		ticker := <-coin.Ticker

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

			// 과거 3분동안만
			minutesCandles, err := upbit.API.GetCandlesMinutes("1", bot.TargetMarket+"-"+coin.Name, "4")
			if err != nil {
				panic(err)
			}

			condition := upbit.API.GetMarketConditionBy(minutesCandles, price)

			// 변동성 돌파는 전략의 기본 조건이다.
			if bot.Predicate("KRW-"+coin.Name, ticker) {
				if coinBalance, ok := balances[coin.Name]; ok {
					// 이미 코인을 가지고 있는 경우

					avgBuyPrice, err := upbit.API.GetAverageBuyPrice(upbit.Accounts, coin.Name)
					if err != nil {
						panic(err)
					}

					pp := price / avgBuyPrice

					// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우, 단기 시장이 하락장일 때 '매도'
					if pp-1 >= p.H && !condition {
						coin.Order("ask", coinBalance, price)

						err := upbit.Db.Update(func(tx *bolt.Tx) error {
							bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

							// 매도한 코인을 다시 추적하지 않도록 설정한다.
							if err := bucket.Put([]byte(coin.Name), []byte{bot.STOPPED}); err != nil {
								return err
							}

							return nil
						})
						if err != nil {
							panic(err)
						}

						continue
					}

					// 가격이 매수평균가 대비 `p.L` 보다 하락한 경우 추가 '매수' 요청
					if pp-1 <= p.L {
						coin.Order("bid", volume, price)
						continue
					}
				} else {
					// 단기 상승장 + 현재 코인을 가지고 있지 않고, 돌파했다면 '매수'
					if condition {
						coin.Order("bid", volume, price)
						continue
					}
				}
			}
		}

		time.Sleep(time.Second * 1)
	}
}
