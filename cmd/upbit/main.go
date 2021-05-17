package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	acc, err := bot.NewFakeAccounts("accounts.db", 55000.0) // 테스트용 계정
	if err != nil {
		logrus.Fatal(err)
	}
	b := &bot.Bot{
		Client:          &client.Client{Client: http.DefaultClient, AccessKey: static.Config.AccessKey, SecretKey: static.Config.SecretKey},
		QuotationClient: &client.QuotationClient{Client: http.DefaultClient},
		Accounts:        acc,
		Strategies:      []bot.Strategy{&bot.Penetration{K: static.Config.K, H: 0.03, L: -0.05}}}

	b.Run()
}
