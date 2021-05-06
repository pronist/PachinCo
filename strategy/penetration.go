package strategy

import (
	"bytes"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"math"
	"strings"
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
	err := upbit.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		markets, err := upbit.API.GetMarkets()
		if err != nil {
			upbit.Logger.Fatal(err)
		}

		K := funk.Chain(markets).
			Map(func(market map[string]interface{}) string {
				return market["market"].(string)
			}).
			Filter(func(market string) bool {
				return strings.HasPrefix(market, bot.Currency)
			})

		K.ForEach(func (market string) {
			ticker, err := upbit.API.GetTicker(market)
			if err != nil {
				upbit.Logger.Fatal(err)
			}

			// 현재 매수/매도를 위해 트래킹 중인 코인이 아니어야 하며
			if bot.Predicate(market, ticker[0]) && bytes.Compare(bucket.Get([]byte(market[4:])), []byte{bot.TRACKING}) != 0 {
				upbit.Logger.WithFields(logrus.Fields{"coin": market[4:]}).Warn("[STRATEGY] PREPARE PENETRATION")

				// 봇 시작시 이미 돌파된 종목에 대해서는 추적을 하지 안도록 한다.
				if err := bucket.Put([]byte(market[4:]), []byte{bot.STOPPED}); err != nil {
					upbit.Logger.Fatal(err)
				}
			}

			time.Sleep(time.Second * 1)
		})

		return nil
	})
	if err != nil {
		upbit.Logger.Panic(err)
	}
}

func (p *Penetration) Run(coin *bot.Coin) {
	err := upbit.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		bot.LogChan <- upbit.Log{
			Msg: "[STRATEGY] PENETRATION",
			Fields: logrus.Fields{
				"coin": coin.Name,
			},
			Level: logrus.WarnLevel,
		}

		for bytes.Compare(bucket.Get([]byte(coin.Name)), []byte{bot.TRACKING}) == 0 {
			ticker := (<-coin.Ticker)[0]

			price := ticker["trade_price"].(float64)

			balances, err := upbit.API.GetBalances(upbit.Accounts)
			if err != nil {
				return err
			}

			if balances["KRW"] >= upbit.MinimumOrderPrice && balances["KRW"] > coin.OnceOrderPrice && coin.OnceOrderPrice > upbit.MinimumOrderPrice {
				volume := coin.OnceOrderPrice / price

				if math.IsInf(volume, 0) {
					return errors.New("division by zero")
				}

				// 과거 3분동안만
				minutesCandles, err := upbit.API.GetCandlesMinutes("1", "KRW"+coin.Name, "4")
				if err != nil {
					return err
				}

				condition := upbit.API.GetMarketConditionBy(minutesCandles, price)

				// 변동성 돌파는 전략의 기본 조건이다.
				if bot.Predicate("KRW"+ coin.Name, ticker) {
					if coinBalance, ok := balances[coin.Name]; ok {
						// 이미 코인을 가지고 있는 경우

						avgBuyPrice, err := upbit.API.GetAverageBuyPrice(upbit.Accounts, coin.Name)
						if err != nil {
							return err
						}

						pp := price / avgBuyPrice

						// 매수 평균가 대비 현재 가격의 '상승률' 이 `p.H` 보다 큰 경우, 단기 시장이 하락장일 때 '매도'
						if pp-1 >= p.H && !condition {
							coin.Order("ask", coinBalance, price)

							// 매도한 코인을 다시 추적하지 않도록 설정한다.
							if err := bucket.Put([]byte(coin.Name), []byte{bot.STOPPED}); err != nil {
								return err
							}
							continue
						}

						// 가격이 매수평균가 대비 `p.L` 보다 하락한 경우 추가 매수 요청
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

			time.Sleep(1 * time.Second)
		}

		return nil
	})
	if err != nil {
		bot.LogChan <- upbit.Log{Msg: err.Error(), Level: logrus.FatalLevel}
	}
}
