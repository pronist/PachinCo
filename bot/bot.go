package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
)

var LogChan = make(chan Log)

type Bot struct {
	Strategy Strategy
}

func New(Strategy Strategy) *Bot {
	// 요청 자체는 최대 동시에 10번 까지 처리할 수 있는 듯하나,
	// 매수/매도를 체크할 때 요청을 두 번하는 경우가 있어서 추적마켓을 줄일 필요가 있음.
	if len(upbit.Config.Coins) > 10 {
		logrus.Fatal("Tracking markets must less than 11")
	}

	return &Bot{Strategy}
}

func (b *Bot) Run() {
	for coin := range upbit.Config.Coins {
		go b.Strategy.Tracking(upbit.Config.Coins, coin)
	}
}

func (b *Bot) StartLogging() {
	for {
		select {
		case log := <-LogChan:
			Logger.WithFields(log.Fields).Log(log.Level, log.Msg)
		}
	}
}
