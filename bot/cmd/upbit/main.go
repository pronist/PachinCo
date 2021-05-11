package main

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/accounts"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/strategy"
)

func main() {
	acc := accounts.NewTestAccounts(55000.0) // 테스트용 계정

	b := &bot.Bot{Accounts: acc, Strategies: []bot.Strategy{
		&strategy.Penetration{K: upbit.Config.K, H: 0.03, L: -0.05},
	}}

	b.Run()
}
