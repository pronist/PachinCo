package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/api"
	"github.com/pronist/upbit/gateway"
	"net/http"
)

func main() {
	config := upbit.NewConfig("upbit.config.yml")

	bot := upbit.NewBot(&upbit.Strategy{Api: &api.API{
		Client: &gateway.Client{AccessKey: config.Bot.AccessKey, SecretKey: config.Bot.SecretKey},
		QuotationClient: &gateway.QuotationClient{Client: &http.Client{}},
	}})
	bot.Run()
}
