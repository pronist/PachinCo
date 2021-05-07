package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/strategy"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			bot.Logger <- bot.Log{Msg: err, Level: logrus.ErrorLevel}
		}
	}()

	b := &bot.Bot{[]bot.Strategy{
		//&strategy.Basic{F: -0.03, L: -0.03, H: 0.03},
		&strategy.Penetration{K: upbit.Config.K, H: 0.05, L: -0.05},
	}}

	b.Run()
}
