package ticker

import (
	"fmt"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	errLogChan = make(chan upbit.Log)
	tickLogChan = make(chan upbit.Log)
)

type Ticker struct {
	config *upbit.Config
	api    *api.API
}

func New() *Ticker {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		logrus.Panic(err)
	}

	return &Ticker{config, &api.API{
		Client:          &client.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &client.QuotationClient{Client: &http.Client{}},
	}}
}

func (t *Ticker) Run() {
	for coin := range t.config.Coins {
		go t.Tick(coin)
	}

	t.Logging()
}

func (t *Ticker) Tick(coin string) {
	for {
		changeRate, err := t.api.GetChangeRate("KRW-" + coin)
		if err != nil {
			errLogChan <- upbit.Log{Msg: err}
		}
		tickLogChan <- upbit.Log{Msg: coin, Fields: logrus.Fields{
			"change-rate": fmt.Sprintf("%.2f%%", changeRate*100),
		}}

		time.Sleep(1 * time.Second)
	}
}

func (t *Ticker) Logging()  {
	errLogger := upbit.NewLogger("logs/error.log", logrus.ErrorLevel, true)
	tickLogger := upbit.NewLogger("", logrus.InfoLevel, false)

	for {
		select {
		case errLog := <-errLogChan: errLogger.WithFields(errLog.Fields).WithError(errLog.Msg.(error))
		case tickLog := <-tickLogChan: tickLogger.WithFields(tickLog.Fields).Info(tickLog.Msg)
		}
	}
}
