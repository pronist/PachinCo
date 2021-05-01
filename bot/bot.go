package bot

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"github.com/sirupsen/logrus"
	"net/http"
)

var exit = make(chan string)

var (
	ErrorLogger = upbit.NewLogger("logs/error.log", logrus.ErrorLevel, true)
	EventLogger = upbit.NewLogger("logs/log.log", logrus.WarnLevel, false)
)

type Bot struct {
	config *upbit.Config
	api    *api.API
}

func New() *Bot {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		logrus.Panic(err)
	}

	return &Bot{config, &api.API{
		Client:          &client.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &client.QuotationClient{Client: &http.Client{}},
	}}
}

func (b *Bot) Run() {
	for coin := range b.config.Coins {
		go b.B(b.config.Coins, coin)
		go b.S(b.config.Coins, coin)
	}

	logrus.Panic(<-exit)
}
