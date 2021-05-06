package bot

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"time"
)

const Currency = "KRW"

const (
	TRACKING = 0x01 // 코인을 트래킹하고 있다
	STOPPED  = 0x02 // 모종의 이유로 인해 추적이 중단된 코인이다.
)

var LogChan = make(chan upbit.Log) // 외부에서 사용하게 될 로그 채널이다.

type Bot struct {
	Strategies []Strategy
}

func (b *Bot) Run() {
	detector := NewDetector()
	go detector.Run(Currency, Predicate) // 종목 찾기 시작!

	err := b.runAlreadyTrackingCoins()
	if err != nil {
		upbit.Logger.Fatal(err)
	}

	for {
		select {
		case ticker := <-detector.D:
			market := ticker["code"].(string)

			coin, err := NewCoin(market[4:], upbit.Config.C)
			if err != nil {
				upbit.Logger.Fatal(err)
			}

			err = upbit.Db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

				// 이미 코인이 담겨져 있다면 추적상태로 바꾸지 않는다.
				if bucket.Get([]byte(coin.Name)) == nil {
					// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
					if err := bucket.Put([]byte(coin.Name), []byte{TRACKING}); err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				upbit.Logger.Fatal(err)
			}

			go b.tick(coin)

			for _, strategy := range b.Strategies {
				go strategy.Run(coin)
			}
		case log := <-LogChan:
			upbit.Logger.WithFields(log.Fields).Log(log.Level, log.Msg)
		}
	}
}

// 데이터베이스에 이미 트래킹 중인 코인이 있다면 전략을 시작하자.
func (b *Bot) runAlreadyTrackingCoins() error {
	err := upbit.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		err := bucket.ForEach(func(k, v []byte) error {
			if bytes.Equal(v, []byte{TRACKING}) {
				coin, err := NewCoin(Currency+"-"+string(k), upbit.Config.C)
				if err != nil {
					return err
				}

				go b.tick(coin)

				for _, strategy := range b.Strategies {
					go strategy.Run(coin)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) tick(coin *Coin) {
	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		upbit.Logger.Fatal(err)
	}

	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": []string{Currency + "-" + coin.Name}, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	for {
		var r map[string]interface{}

		if err := ws.WriteJSON(data); err != nil {
			upbit.Logger.Fatal(err)
		}

		if err := ws.ReadJSON(&r); err != nil {
			upbit.Logger.Fatal(err)
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			coin.Ticker <- r
		}

		time.Sleep(time.Second * 1)
	}
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