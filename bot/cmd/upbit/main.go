package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/strategy"
)

func main() {
	bot := bot.New(&strategy.Basic{F: -0.03, L: -0.03, H: 0.03})

	bot.Run()
	bot.StartLogging()
}
