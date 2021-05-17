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
	logger <- log{msg: "Bot started...", level: logrus.DebugLevel}

	// 전략의 사전 준비를 해야한다.
	//for _, strategy := range b.Strategies {
	//	strategy.Prepare(b.Accounts)
	//}

	///// 이미 가지고 있는 코인에 대해서는 전략을 시작해야 한다.
	err := b.runStrategyForCoinsInHands()
	if err != nil {
		panic(err)
	}
	/////

	///// 디텍터
	detector, err := newDetector()
	if err != nil {
		panic(err)
	}

	go detector.run(b, Market, predicate) // 종목 찾기 시작!
	/////

	for {
		select {
		// 디텍팅되어 가져온 코인에 대해서 전략 시작 ...
		case tick := <-detector.d:
			market := tick["code"].(string)

			if _, ok := MarketTrackingStates[market]; !ok && Tracking < Config.Max {
				//
				logger <- log{
					msg: "Detected",
					fields: logrus.Fields{
						"market":      market,
						"change-rate": tick["signed_change_rate"].(float64),
						"price":       tick["trade_price"].(float64),
					},
					level: logrus.DebugLevel,
				}
				//
				if err := b.launch(market); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (b *Bot) runStrategyForCoinsInHands() error {
	acc, err := b.Accounts.accounts()
	if err != nil {
		return err
	}

	balances := getBalances(acc)
	delete(balances, "KRW")

	for coin := range balances {
		if err := b.launch(Market + "-" + coin); err != nil {
			return err
		}
	}
	//
	logger <- log{msg: "Run strategy for coins in hands.", level: logrus.DebugLevel}
	//
	return nil
}

func (b *Bot) launch(market string) error {
	// 코인 생성
	coin, err := newCoin(b.Accounts, market[4:], Config.C)
	if err != nil {
		return err
	}

	// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
	MarketTrackingStates[market] = TRACKING
	Tracking++

	// 전략에 주기적으로 가격 정보를 보낸다.
	go b.tick(coin)

	for _, strategy := range b.Strategies {
		go b.Strategy(coin, strategy)
	}

	return nil
}

func (b *Bot) Strategy(c *coin, strategy Strategy) {
	defer func() {
		if err := recover(); err != nil {
			//
			logger <- log{
				msg: err,
				fields: logrus.Fields{
					"role": "Strategy", "strategy": reflect.TypeOf(strategy).Name(), "coin": c.name,
				},
				level: logrus.ErrorLevel,
			}
			//
		}
	}()
	//
	logger <- log{
		msg:    "STARTED",
		fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": c.name},
		level:  logrus.DebugLevel,
	}
	//
	stat, ok := MarketTrackingStates[Market+"-"+c.name]

	for ok && stat == TRACKING {
		t := <-c.t

		acc, err := b.Accounts.accounts()
		if err != nil {
			panic(err)
		}

		balances := getBalances(acc)
		if balances["KRW"] >= MinimumOrderPrice && balances["KRW"] > c.onceOrderPrice && c.onceOrderPrice > MinimumOrderPrice {
			if _, err := strategy.run(b.Accounts, c, t); err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Second * 1)
	}

	//
	logger <- log{
		msg:    "CLOSED",
		fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": c.name},
		level:  logrus.DebugLevel,
	}
	//
}

func (b *Bot) tick(c *coin) {
	defer func() {
		if err := recover(); err != nil {
			logger <- log{msg: err, fields: logrus.Fields{"role": "Tick", "coin": c.name}, level: logrus.ErrorLevel}
		}
	}()

	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		panic(err)
	}

	m := Market + "-" + c.name

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

		logger <- log{
			msg: c.name,
			fields: logrus.Fields{
				"change-rate": r["signed_change_rate"].(float64),
				"price":       r["trade_price"].(float64),
			},
			level: logrus.TraceLevel,
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			c.t <- r
		}

		time.Sleep(time.Second * 1)
	}

	//
	logger <- log{
		msg:    "CLOSED",
		fields: logrus.Fields{"role": "Tick", "coin": c.name},
		level:  logrus.DebugLevel,
	}
	//
}
