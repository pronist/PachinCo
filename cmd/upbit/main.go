package main

import (
	"github.com/pronist/upbit"
	"log"
	"net/http"
)

const (
	accessKey = "UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4"
	secretKey = "2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn"
)

func main() {
	bot, err := upbit.NewBot(
		&upbit.Client{AccessKey: accessKey, SecretKey: secretKey},
		&upbit.QuotationClient{Client: &http.Client{}})
	if err != nil {
		log.Panic(err)
	}

	bot.Buy("BTT")
}
