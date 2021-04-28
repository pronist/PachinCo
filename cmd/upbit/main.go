package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/api"
	"github.com/pronist/upbit/gateway"
	"github.com/pronist/upbit/strategy"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		logrus.Panic(err)
	}

	upDownStrategy := strategy.UpDown{F: -0.03, L: -0.03, H: 0.03}

	bot := upbit.NewBot(&upDownStrategy, &api.API{
		Client:          &gateway.Client{AccessKey: config.Bot.AccessKey, SecretKey: config.Bot.SecretKey},
		QuotationClient: &gateway.QuotationClient{Client: &http.Client{}},
	})
	bot.Run()
}
