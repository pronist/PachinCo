package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/strategy"
)

func main() {
	bot := &bot.Bot{[]bot.Strategy{
		&strategy.Basic{F: -0.03, L: -0.03, H: 0.03},
		//&strategy.Penetration{Strategy: &strategy.Strategy{}, K: 0.5},
	}}

	bot.Run()
}
