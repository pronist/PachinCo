package upbit

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

const MinimumOrderPrice = 5000 // 업비트의 최소 매도/매수 가격은 '5000 KRW'
const Market = "KRW"           // 원화 마켓을 추적한다.

const (
	B = "bid" // 매수
	S = "ask" // 매도
)

// 추적 상태를 나타내는 상수들
const (
	TRACKING = iota
	STOPPED
)

var MarketTrackingStates = make(map[string]int)
var Tracking int

// 트래킹할 종목에 대한 조건이다.
func Predicate(b *Bot, t map[string]interface{}) bool {
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

	return dayCandles[0]["opening_price"].(float64)+(R*Config.K) < price
}

//func Predicate(t map[string]interface{}) bool {
//	return true
//}

type Bot struct {
	*Client
	*QuotationClient
	Accounts   Accounts
	Strategies []Strategy
}

func (b *Bot) Run() {
	Logger <- Log{Msg: "Bot started...", Level: logrus.DebugLevel}

	// 전략의 사전 준비를 해야한다.
	//for _, strategy := range b.Strategies {
	//	strategy.Prepare(b.Accounts)
	//}

	///// 이미 가지고 있는 코인에 대해서는 전략을 시작해야 한다.
	err := b.RunStrategyForCoinsInHands()
	if err != nil {
		panic(err)
	}
	/////

	///// 디텍터
	detector, err := NewDetector()
	if err != nil {
		panic(err)
	}

	go detector.Run(b, Market, Predicate) // 종목 찾기 시작!
	/////

	for {
		select {
		// 디텍팅되어 가져온 코인에 대해서 전략 시작 ...
		case tick := <-detector.D:
			market := tick["code"].(string)

			if _, ok := MarketTrackingStates[market]; !ok && Tracking < Config.Max {
				//
				Logger <- Log{
					Msg: "Detected",
					Fields: logrus.Fields{
						"market":      market,
						"change-rate": tick["signed_change_rate"].(float64),
						"price":       tick["trade_price"].(float64),
					},
					Level: logrus.DebugLevel,
				}
				//
				if err := b.Go(market); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (b *Bot) RunStrategyForCoinsInHands() error {
	acc, err := b.Accounts.Accounts()
	if err != nil {
		return err
	}

	balances := GetBalances(acc)
	delete(balances, "KRW")

	for coin := range balances {
		if err := b.Go(Market + "-" + coin); err != nil {
			return err
		}
	}
	//
	Logger <- Log{Msg: "Run strategy for coins in hands.", Level: logrus.DebugLevel}
	//
	return nil
}

func (b *Bot) Go(market string) error {
	// 코인 생성
	coin, err := NewCoin(b.Accounts, market[4:], Config.C)
	if err != nil {
		return err
	}

	// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
	MarketTrackingStates[market] = TRACKING
	Tracking++

	// 전략에 주기적으로 가격 정보를 보낸다.
	go b.Tick(coin)

	for _, strategy := range b.Strategies {
		go b.Strategy(coin, strategy)
	}

	return nil
}

func (b *Bot) Strategy(coin *Coin, strategy Strategy) {
	defer func(coin *Coin) {
		if err := recover(); err != nil {
			//
			Logger <- Log{
				Msg: err,
				Fields: logrus.Fields{
					"role": "Strategy", "strategy": reflect.TypeOf(strategy).Name(), "coin": coin.Name,
				},
				Level: logrus.ErrorLevel,
			}
			//
		}
	}(coin)
	//
	Logger <- Log{
		Msg:    "STARTED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": coin.Name},
		Level:  logrus.DebugLevel,
	}
	//
	stat, ok := MarketTrackingStates[Market+"-"+coin.Name]

	for ok && stat == TRACKING {
		t := <-coin.T

		acc, err := b.Accounts.Accounts()
		if err != nil {
			panic(err)
		}

		balances := GetBalances(acc)
		if balances["KRW"] >= MinimumOrderPrice && balances["KRW"] > coin.OnceOrderPrice && coin.OnceOrderPrice > MinimumOrderPrice {
			if _, err := strategy.Run(b.Accounts, coin, t); err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Second * 1)
	}

	//
	Logger <- Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": coin.Name},
		Level:  logrus.DebugLevel,
	}
	//
}

func (b *Bot) Tick(coin *Coin) {
	defer func() {
		if err := recover(); err != nil {
			Logger <- Log{Msg: err, Fields: logrus.Fields{"role": "Tick", "coin": coin.Name}, Level: logrus.ErrorLevel}
		}
	}()

	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		panic(err)
	}

	m := Market + "-" + coin.Name

	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": []string{m}, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	for MarketTrackingStates[m] == TRACKING {
		var r map[string]interface{}

		if err := ws.WriteJSON(data); err != nil {
			panic(err)
		}

		if err := ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		Logger <- Log{
			Msg: coin.Name,
			Fields: logrus.Fields{
				"change-rate": r["signed_change_rate"].(float64),
				"price":       r["trade_price"].(float64),
			},
			Level: logrus.TraceLevel,
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			coin.T <- r
		}

		time.Sleep(time.Second * 1)
	}

	//
	Logger <- Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"role": "Tick", "coin": coin.Name},
		Level:  logrus.DebugLevel,
	}
	//
}
