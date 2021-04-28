package upbit

import (
	"github.com/pronist/upbit/api"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	config *config
	api    *api.API

	// 고루틴에서 보고하기 위한 각종 채널
	ticker  chan Log
	err     chan Log
	logging chan Log
}

func NewBot(config *config, api *api.API) *Bot {
	return &Bot{config, api, make(chan Log), make(chan Log), make(chan Log)}
}

func (b *Bot) Run() {
	for coin := range b.config.Coins {
		go b.Watch(coin)
		go b.B(b.config.Coins, coin)
		go b.S(b.config.Coins, coin)
	}
}

func (b *Bot) Logging() {
	errLogger, err := NewLogger("logs/error.log", logrus.ErrorLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	logLogger, err := NewLogger("logs/log.log", logrus.WarnLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	stdLogger, err := NewLogger("", logrus.InfoLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	for {
		select {
		case t := <-b.ticker:
			stdLogger.WithFields(t.Fields).Info(t.Msg)
		case e := <-b.err:
			if e.Terminate {
				errLogger.WithFields(e.Fields).Fatal(e.Msg)
			} else {
				errLogger.WithFields(e.Fields).Error(e.Msg)
			}
		case l := <-b.logging:
			logLogger.WithFields(l.Fields).Warn(l.Msg)
		}
	}
}
