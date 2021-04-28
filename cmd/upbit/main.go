package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/api"
	"github.com/pronist/upbit/gateway"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		logrus.Panic(err)
	}

	bot := upbit.NewBot(config, &api.API{
		Client:          &gateway.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &gateway.QuotationClient{Client: &http.Client{}},
	})
	bot.Run()
	bot.Logging()
}
