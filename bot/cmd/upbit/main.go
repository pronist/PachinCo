package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/strategy"
)

func main() {
	bot := bot.New(&strategy.Basic{})

	bot.Run()
	bot.StartLogging()
}
