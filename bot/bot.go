package bot

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"time"
)

const Currency = "KRW"

const (
	TRACKING = 0x01 // 코인을 트래킹하고 있다
	STOPPED = 0X02 // 모종의 이유로 인해 추적이 중단된 코인이다.
)

var LogChan = make(chan upbit.Log) // 외부에서 사용하게 될 로그 채널이다.

type Bot struct {
	Strategies []Strategy
}

// 트래킹할 종목에 대한 조건이다.
func Predicate(market string, r map[string]interface{}) bool {
	price := r["trade_price"].(float64)

	// https://wikidocs.net/21888
	dayCandles, err := upbit.API.GetCandlesDays(market, "2")
	if err != nil {
		LogChan <- upbit.Log{Msg: err, Level: logrus.ErrorLevel}
	}

	// "변동성 돌파" 한 종목을 트래킹할 조건으로 설정.
	R := dayCandles[1]["high_price"].(float64) - dayCandles[1]["low_price"].(float64)

	return dayCandles[0]["opening_price"].(float64)+(R*upbit.Config.K) < price
}

func (b *Bot) Run() {
	detector := NewDetector()
	go detector.Search(Currency, Predicate) // 종목 찾기 시작!

	err := upbit.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		for {
			select {
			case ticker := <-detector.D:
				market := ticker["code"].(string)

				// 새롭게 발견된 코인을 생성하고 데이터베이스에 추가한다.
				if coin, err := NewCoin(market[4:], upbit.Config.C); err == nil {
					// 전략마다 틱을 별도로 실행하면 요청횟수 제한이 걸린다.
					go b.tick(coin)

					// 봇이 실행할 전략을 코인을 대상으로 실행한다.
					for _, strategy := range b.Strategies {
						go strategy.Run(coin)
					}

					// 이미 코인이 중지 상태로 담겨져 있다면 추적상태로 바꾸지 않는다.
					if bytes.Compare(bucket.Get([]byte(market[4:])), []byte{STOPPED}) != 0 {
						// 코인을 추척하고 있다는 것을 저장해둔다.
						// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
						if err := bucket.Put([]byte(market[4:]), []byte{TRACKING}); err != nil {
							return err
						}
					}
				}
			case log := <-LogChan:
				upbit.Logger.WithFields(log.Fields).Log(log.Level, log.Msg)
			}
		}
	})
	if err != nil {
		upbit.Logger.Panic(err)
	}
}

func (b *Bot) tick(coin *Coin) {
	for {
		ticker, err := upbit.API.GetTicker("KRW-" + coin.Name)
		if err != nil {
			LogChan <- upbit.Log{Msg: err, Level: logrus.ErrorLevel}
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			coin.Ticker <- ticker
		}

		time.Sleep(time.Second * 1)
	}
}
