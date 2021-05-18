package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
)

func main() {
	acc, err := bot.NewFakeAccounts("accounts.db", 55000.0) // 테스트용 계정
	if err != nil {
		logrus.Fatal(err)
	}

	b := bot.NewBot(acc, []bot.Strategy{
		&bot.Penetration{K: static.Config.K, H: 0.03, L: -0.05},
	})

	if err := b.Run(); err != nil {
		logrus.Fatal(err)
	}
}
