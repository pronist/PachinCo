package main

import (
	"github.com/pronist/upbit/bot"
	"github.com/sirupsen/logrus"
)

func main() {
	///// 봇에 사용할 전략을 설정한다.
	b := bot.New([]bot.Strategy{
		// https://wikidocs.net/21888
		&bot.PenetrationStrategy{},
	})
	/////

	///// 봇에 사용할 계정을 설정한다.
	//acc, err := bot.NewUpbitAccounts(b)
	acc, err := bot.NewFakeAccounts("accounts.db", 55000.0) // 테스트용 계정
	if err != nil {
		logrus.Fatal(err)
	}

	b.SetAccounts(acc)
	/////

	logrus.Panic(b.Run())
}
