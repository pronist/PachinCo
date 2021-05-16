package main

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	acc, err := upbit.NewFakeAccounts(upbit.FaccDbName, 55000.0) // 테스트용 계정
	if err != nil {
		upbit.Logger <- upbit.Log{Msg: err, Level: logrus.FatalLevel}
	}

	b := &upbit.Bot{
		Client:          &upbit.Client{Client: http.DefaultClient, AccessKey: upbit.Config.AccessKey, SecretKey: upbit.Config.SecretKey},
		QuotationClient: &upbit.QuotationClient{Client: http.DefaultClient},
		Accounts:        acc,
		Strategies:      []upbit.Strategy{&upbit.Penetration{K: upbit.Config.K, H: 0.03, L: -0.05}}}

	b.Run()
}
