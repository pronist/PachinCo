package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/accounts"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/strategy"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
		}
	}()

	acc := accounts.NewTestAccounts("5000000.0") // 테스트용 계정

	b := &bot.Bot{acc, []bot.Strategy{
		&strategy.Penetration{K: upbit.Config.K, H: 0.03, L: -0.05},
	}}

	b.Run()
}
