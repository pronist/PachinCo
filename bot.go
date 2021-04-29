package upbit

import (
	"github.com/pronist/upbit/api"
	"github.com/pronist/upbit/gateway"
	"github.com/sirupsen/logrus"
	"net/http"
)

var Exit = make(chan string)

type Bot struct {
	config *config
	api    *api.API
}

func NewBot() *Bot {
	config, err := newConfig("upbit.config.yml")
	if err != nil {
		Exit <- err.Error()
	}

	return &Bot{config, &api.API{
		Client:          &gateway.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &gateway.QuotationClient{Client: &http.Client{}},
	}}
}

func (b *Bot) Run() {
	for coin := range b.config.Coins {
		go b.Watch(coin)
		go b.B(b.config.Coins, coin)
		go b.S(b.config.Coins, coin)
	}

	logrus.Panic(<-Exit)
}