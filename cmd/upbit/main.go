package main

import (
	"github.com/pronist/upbit"
	"log"
	"net/http"
)

func main() {
	config := upbit.NewConfig("upbit.config.yml")

	strategy, err := upbit.NewStrategy(
		&upbit.Client{AccessKey: config.Bot.AccessKey, SecretKey: config.Bot.SecretKey},
		&upbit.QuotationClient{Client: &http.Client{}})
	if err != nil {
		log.Panic(err)
	}

	bot := upbit.NewBot(strategy)
	bot.Run()
}
