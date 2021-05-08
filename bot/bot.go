package bot

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/log"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"time"
)

const TargetMarket = "KRW"

// 추적 상태를 나타내는 상수들
const (
	TRACKING = iota
	STOPPED
)

// 트래킹할 종목에 대한 조건이다.
func Predicate(market string, r map[string]interface{}) bool {
	price := r["trade_price"].(float64)

	// https://wikidocs.net/21888
	dayCandles, err := upbit.API.GetCandlesDays(market, "2")
	if err != nil {
		panic(err)
	}

	// "변동성 돌파" 한 종목을 트래킹할 조건으로 설정.
	R := dayCandles[1]["high_price"].(float64) - dayCandles[1]["low_price"].(float64)

	return dayCandles[0]["opening_price"].(float64)+(R*upbit.Config.K) < price
}

type Bot struct {
	Strategies []Strategy
}

func (b *Bot) Run() {
	log.Logger <- log.Log{Msg: "Bot started...", Level: logrus.InfoLevel}

	// 전략의 사전 준비를 해야한다.
	for _, strategy := range b.Strategies {
		strategy.Prepare()
	}

	///// 디텍터
	detector, err := NewDetector()
	if err != nil {
		panic(err)
	}

	go detector.Run(TargetMarket, Predicate) // 종목 찾기 시작!
	/////

	///// 이미 가지고 있는 코인에 대해서는 전략을 시작해야 한다.
	err = b.RunStrategyForCoinsInHands()
	if err != nil {
		panic(err)
	}
	/////

	for {
		select {
		// 디텍팅되어 가져온 코인에 대해서 전략 시작 ...
		case tick := <-detector.D:
			market := tick["code"].(string)

			log.Logger <- log.Log{
				Msg: "Detected",
				Fields: logrus.Fields{
					"coin":        market[4:],
					"change-rate": tick["change_rate"].(float64),
					"price":       tick["trade_price"].(float64),
				},
				Level: logrus.InfoLevel,
			}

			if err := b.Go(market[4:]); err != nil {
				panic(err)
			}
		}
	}
}

func (b *Bot) RunStrategyForCoinsInHands() error {
	balances, err := upbit.API.GetBalances(upbit.Accounts)
	if err != nil {
		return err
	}
	delete(balances, "KRW")

	for coin := range balances {
		if err := b.Go(coin); err != nil {
			return err
		}
	}

	log.Logger <- log.Log{Msg: "Run strategy for coins in hands.", Level: logrus.InfoLevel}

	return nil
}

func (b *Bot) Go(coinName string) error {
	coin, err := NewCoin(coinName, upbit.Config.C)
	if err != nil {
		return err
	}

	err = upbit.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(upbit.CoinsBucketName))

		// 이미 코인이 담겨져 있다면 추적상태로 바꾸지 않는다.
		if bucket.Get([]byte(coin.Name)) == nil {
			// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
			if err := bucket.Put([]byte(coin.Name), []byte{TRACKING}); err != nil {
				return err
			}

			log.Logger <- log.Log{
				Msg: "Tracking starts with",
				Fields: logrus.Fields{
					"coin": coin.Name,
				},
				Level: logrus.WarnLevel,
			}

			go b.Tick(coin)

			for _, strategy := range b.Strategies {
				go strategy.Run(coin)
			}
		}

		return nil
	})

	return err
}

func (b *Bot) Tick(coin *Coin) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger <- log.Log{Msg: err, Fields: logrus.Fields{"role": "Tick", "coin": coin.Name}, Level: logrus.ErrorLevel}
		}
	}()

	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		panic(err)
	}

	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": []string{TargetMarket + "-" + coin.Name}, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	for {
		var r map[string]interface{}

		if err := ws.WriteJSON(data); err != nil {
			panic(err)
		}

		if err := ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			coin.T <- r
		}

		time.Sleep(time.Second * 1)
	}
}
