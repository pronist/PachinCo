package bot

import (
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"time"
)

type Bot struct {
	*client.Client
	*client.QuotationClient
	Accounts   Accounts
	Strategies []Strategy
}

// 새로운 봇을 만든다. 봇에 사용할 계정과 전략을 받는다.
func NewBot(accounts Accounts, strategies []Strategy) *Bot {
	c := &client.Client{
		Client:    http.DefaultClient,
		AccessKey: static.Config.AccessKey,
		SecretKey: static.Config.SecretKey,
	}
	qc := &client.QuotationClient{Client: http.DefaultClient}

	return &Bot{c, qc, accounts, strategies}
}

// 봇을 실행한다. 다음과 같은 일이 발생한다.
//
// 1. 계좌를 조회하여 현재 가지고 있는 코인에 대해 틱과 전략를 실행한다.
// 2. 디텍터를 실행하여 predicate 에 부합하는 종목을 탐색하여 보고, 탐색된 종목에 대해 틱과 전략를 실행한다.
func (b *Bot) Run() error {
	log.Logger <- log.Log{Msg: "Bot started...", Level: logrus.DebugLevel}

	// 전략의 사전 준비를 해야한다.
	for _, strategy := range b.Strategies {
		strategy.prepare(b)
	}

	///// 이미 가지고 있는 코인에 대해서는 전략을 시작해야 한다.
	if err := b.inHands(); err != nil {
		return err
	}
	/////

	///// 디텍터
	detector, err := newDetector()
	if err != nil {
		return err
	}

	go detector.run(b, predicate) // 종목 찾기 시작!
	/////

	for {
		select {
		// 디텍팅되어 가져온 코인에 대해서 전략 시작 ...
		case tick := <-detector.d:
			market := tick["code"].(string)

			if _, ok := stat[market]; !ok {
				//
				log.Logger <- log.Log{
					Msg: "Detected",
					Fields: logrus.Fields{
						"market":      market,
						"change-rate": tick["signed_change_rate"].(float64),
						"price":       tick["trade_price"].(float64),
					},
					Level: logrus.DebugLevel,
				}
				//
				if err := b.launch(market); err != nil {
					return err
				}
			}
		}
	}
}

func (b *Bot) inHands() error {
	acc, err := b.Accounts.accounts()
	if err != nil {
		return err
	}

	balances := getBalances(acc)
	delete(balances, "KRW")

	for coin := range balances {
		if err := b.launch(targetMarket + "-" + coin); err != nil {
			return err
		}
	}
	//
	log.Logger <- log.Log{Msg: "Run strategy for coins in hands.", Level: logrus.DebugLevel}
	//
	return nil
}

func (b *Bot) launch(market string) error {
	// 코인 생성
	coin, err := newCoin(b.Accounts, market[4:], static.Config.C)
	if err != nil {
		return err
	}

	// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
	stat[market] = tracked

	// 전략에 주기적으로 가격 정보를 보낸다.
	go b.tick(coin)

	for _, strategy := range b.Strategies {
		go b.strategy(coin, strategy)
	}

	return nil
}

func (b *Bot) strategy(c *coin, strategy Strategy) {
	defer func() {
		if err := recover(); err != nil {
			//
			log.Logger <- log.Log{
				Msg: err,
				Fields: logrus.Fields{
					"role": "Strategy", "strategy": reflect.TypeOf(strategy).Name(), "coin": c.name,
				},
				Level: logrus.ErrorLevel,
			}
			//
		}
	}()
	//
	log.Logger <- log.Log{
		Msg:    "STARTED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": c.name},
		Level:  logrus.DebugLevel,
	}
	//
	stat, ok := stat[targetMarket+"-"+c.name]

	for ok && stat == tracked {
		t := <-c.t

		acc, err := b.Accounts.accounts()
		if err != nil {
			panic(err)
		}

		balances := getBalances(acc)

		if balances["KRW"] >= minimumOrderPrice && balances["KRW"] > c.onceOrderPrice && c.onceOrderPrice > minimumOrderPrice {
			if _, err := strategy.run(b, c, t); err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Second * 1)
	}

	//
	log.Logger <- log.Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).Name(), "coin": c.name},
		Level:  logrus.DebugLevel,
	}
	//
}

func (b *Bot) tick(c *coin) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger <- log.Log{Msg: err, Fields: logrus.Fields{"role": "Tick", "coin": c.name}, Level: logrus.ErrorLevel}
		}
	}()

	m := targetMarket + "-" + c.name

	wsc, err := client.NewWebsocketClient("ticker", []string{m}, true, false)
	if err != nil {
		panic(err)
	}

	for stat[m] == tracked {
		var r map[string]interface{}

		if err := wsc.Ws.WriteJSON(wsc.Data); err != nil {
			panic(err)
		}

		if err := wsc.Ws.ReadJSON(&r); err != nil {
			panic(err)
		}
		//
		log.Logger <- log.Log{
			Msg: c.name,
			Fields: logrus.Fields{
				"change-rate": r["signed_change_rate"].(float64),
				"price":       r["trade_price"].(float64),
			},
			Level: logrus.TraceLevel,
		}
		//
		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.Strategies {
			c.t <- r
		}

		time.Sleep(time.Second * 1)
	}
	if err := wsc.Ws.Close(); err != nil {
		panic(err)
	}
	//
	log.Logger <- log.Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"role": "Tick", "coin": c.name},
		Level:  logrus.DebugLevel,
	}
	//
}
