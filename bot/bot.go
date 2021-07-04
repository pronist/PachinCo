package bot

import (
	"net/http"
	"reflect"
	"time"

	"github.com/thoas/go-funk"

	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	*client.Client
	*client.QuotationClient
	accounts   Accounts
	strategies []Strategy
}

// 새로운 봇을 만든다. 봇에 사용할 계정과 전략을 받는다.
func New(strategies []Strategy) *Bot {
	c := &client.Client{
		Client:    http.DefaultClient,
		AccessKey: static.Config.KeyPair.AccessKey,
		SecretKey: static.Config.KeyPair.SecretKey,
	}
	qc := &client.QuotationClient{Client: http.DefaultClient}

	return &Bot{c, qc, nil, strategies}
}

// 봇에서 사용할 계정을 설정한다.
func (b *Bot) SetAccounts(accounts Accounts) {
	b.accounts = accounts
}

// 봇을 실행한다. 다음과 같은 일이 발생한다.
//
// 1. 계좌를 조회하여 현재 가지고 있는 코인에 대해 틱과 전략를 실행한다.
// 2. 디텍터를 실행하여 predicate 에 부합하는 종목을 탐색하여 보고, 탐색된 종목에 대해 틱과 전략를 실행한다.
func (b *Bot) Run() error {
	log.Logger <- log.Log{Msg: "Bot started...", Level: logrus.WarnLevel}

	// 전략의 사전 준비를 해야한다.
	for _, strategy := range b.strategies {
		log.Logger <- log.Log{
			Msg:    "Register strategy...",
			Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).String()},
			Level:  logrus.DebugLevel,
		}
		if err := strategy.register(b); err != nil {
			return err
		}
	}

	// 이미 가지고 있는 코인에 대해서는 전략을 시작해야 한다.
	if err := b.inHands(); err != nil {
		return err
	}

	d := newDetector()
	go d.run(b, predicate) // 종목 찾기 시작!

	for tick := range d.d {
		// 디텍팅되어 가져온 코인에 대해서 틱과 전략 시작 ...
		market := tick["code"].(string)

		if b.trackable(market) {
			log.Logger <- log.Log{
				Msg:    "Detected",
				Fields: logrus.Fields{"market": market},
				Level:  logrus.DebugLevel,
			}
			if err := b.launch(market); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Bot) trackable(market string) bool {
	// 최대 추적 마켓이 정해져 있는 경우
	if static.Config.MaxTrackedMarket > 0 {
		if len(getMarketsFromStates(staged)) >= static.Config.MaxTrackedMarket {
			return false
		}
	}

	excluded := funk.Contains(static.Config.Blacklist, market)
	included := funk.Contains(static.Config.Whitelist, market)

	// 블랙리스트에 등록되어 있으면 제외
	if excluded {
		return false
	}

	// 화이트리스트가 없거나, 또는 화이트리스트가 존재하여 포함된 경우
	if len(static.Config.Whitelist) < 1 || included {
		if s, ok := states[market]; ok {
			return s == untracked // 상태가 'untracked' 상태면 가능
		} else {
			return true // 상태에 등록되어 있지 않더라도 가능
		}
	}

	return false
}

func (b *Bot) inHands() error {
	acc, err := b.accounts.accounts()
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
	log.Logger <- log.Log{Msg: "Run strategy for coins in hands.", Level: logrus.DebugLevel}

	return nil
}

func (b *Bot) launch(market string) error {
	// 코인 생성
	coin, err := newCoin(b.accounts, market[4:], static.Config.TradableBalanceRatio)
	if err != nil {
		return err
	}

	// 여기서 담아둔 값은 별도의 고루틴에서 돌고 있는 전략의 실행 여부를 결정하게 된다.
	states[market] = staged

	// 전략에 주기적으로 가격 정보를 보낸다.
	go b.tick(coin)

	for _, strategy := range b.strategies {
		if err := strategy.boot(b, coin); err != nil {
			return err
		}
		log.Logger <- log.Log{
			Msg:    "Booting strategy...",
			Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).String()},
			Level:  logrus.DebugLevel,
		}
		go b.strategy(coin, strategy)
	}

	return nil
}

func (b *Bot) strategy(c *coin, strategy Strategy) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger <- log.Log{
				Msg: err,
				Fields: logrus.Fields{
					"role": "Strategy", "strategy": reflect.TypeOf(strategy).String(), "coin": c.name,
				},
				Level: logrus.ErrorLevel,
			}
		}
	}()

	log.Logger <- log.Log{
		Msg:    "STARTED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).String(), "coin": c.name},
		Level:  logrus.DebugLevel,
	}

	stat, ok := states[targetMarket+"-"+c.name]

	for ok && stat == staged {
		t := <-c.t

		acc, err := b.accounts.accounts()
		if err != nil {
			panic(err)
		}

		balances := getBalances(acc)

		if balances["KRW"] >= minimumOrderPrice && balances["KRW"] > c.onceOrderPrice && c.onceOrderPrice > minimumOrderPrice {
			if _, err := strategy.run(b, c, t); err != nil {
				panic(err)
			}
		}
	}

	log.Logger <- log.Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"strategy": reflect.TypeOf(strategy).String(), "coin": c.name},
		Level:  logrus.DebugLevel,
	}
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

	for states[m] == staged {
		var r map[string]interface{}

		if err := wsc.Ws.WriteJSON(wsc.Data); err != nil {
			panic(err)
		}

		if err := wsc.Ws.ReadJSON(&r); err != nil {
			panic(err)
		}

		log.Logger <- log.Log{
			Msg: c.name,
			Fields: logrus.Fields{
				"change-rate": r["signed_change_rate"].(float64),
				"price":       r["trade_price"].(float64),
			},
			Level: logrus.TraceLevel,
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range b.strategies {
			c.t <- r
		}

		time.Sleep(time.Second * 1)
	}
	if err := wsc.Ws.Close(); err != nil {
		panic(err)
	}

	log.Logger <- log.Log{
		Msg:    "CLOSED",
		Fields: logrus.Fields{"role": "Tick", "coin": c.name},
		Level:  logrus.DebugLevel,
	}
}
