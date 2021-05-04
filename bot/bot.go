package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"sync"
)

var LogChan = make(chan Log) // 외부에서 사용하게 될 로그 채널이다.
var Mu = &sync.Mutex{}

// 계정은 글로벌하게 사용되며 갱신시에는 메모리 동기화를 해야한다.
var Accounts []map[string]interface{}

func init() {
	var err error

	Accounts, err = upbit.API.NewAccounts()
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
}

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

		bot.coins = append(bot.coins, coin)
	}

	return &bot
}

func (b *Bot) Run() {
	for _, coin := range b.coins {
		go coin.Tick() // 코인에 대해 Tick 를 발신한다.

		for _, strategy := range coin.Strategies {
			go strategy.Run(coin)
		}
	}
	for {
		select {
		case log := <-LogChan:
			Logger.WithFields(log.Fields).Log(log.Level, log.Msg)
		}
	}
}
