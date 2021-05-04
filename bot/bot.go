package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)


// 계정은 글로벌하게 사용되며 갱신시에는 메모리 동기화를 해야한다.
var Accounts []map[string]interface{}

func init() {
	var err error

	Accounts, err = upbit.API.NewAccounts()
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
}

var Mutex = &sync.Mutex{}

type Bot struct {
	coins []*Coin
}

func New(strategies []Strategy) *Bot {
	bot := Bot{}

	// 요청 자체는 최대 동시에 10번 까지 처리할 수 있는 듯하나,
	// 매수/매도를 체크할 때 요청을 두 번하는 경우가 있어서 추적마켓을 줄일 필요가 있음.
	if len(upbit.Config.Coins) > 10 {
		logrus.Fatal("Tracking markets must less than 11")
	}

	for name, rate := range upbit.Config.Coins {
		coin, err := NewCoin(name, rate, strategies)
		if err != nil {
			logrus.Fatal(err)
		}

		coin.Attach(&bot)
		bot.coins = append(bot.coins, coin)
	}

	return &bot
}

func (b *Bot) Run() {
	for _, coin := range b.coins {
		for _, strategy := range coin.Strategies {
			go strategy.Run(coin)
		}
	}
}

var LogChan = make(chan Log) // 외부에서 사용하게 될 로그 채널이다.

func (b *Bot) StartLogging() {
	for {
		select {
		case log := <-LogChan:
			Logger.WithFields(log.Fields).Log(log.Level, log.Msg)
		}
	}
}

// 어찌되었든 결국 주문은 봇이 한다.
func (b *Bot) Update(coin *Coin, side string, volume, price float64) {
	b.Order(coin.Name, side, volume, price)
	coin.Refresh()
}

// order 메서드는 주문을 하되 Config.Timeout 만큼이 지나가면 주문을 자동으로 취소한다.
// 매수/매도에 둘다 사용한다.
func (b *Bot) Order(coin string, side string, volume, price float64) {
	uuid, err := upbit.API.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
	LogChan <- Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": "KRW-" + coin, "volume": volume, "price": price,
		},
		Level: logrus.WarnLevel,
	}

	done := make(chan int)

	timer := time.NewTimer(time.Second * upbit.Config.Timeout)

	go upbit.API.Wait(done, uuid)

	select {
	// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
	case <-timer.C:
		err := upbit.API.CancelOrder(uuid)
		if err != nil {
			LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
		}
		LogChan <- Log{
			Msg: "CANCEL",
			Fields: logrus.Fields{
				"coin": coin, "side": side, "timeout": time.Second * upbit.Config.Timeout,
			},
			Level: logrus.WarnLevel,
		}
	case <-done:
	}
}
