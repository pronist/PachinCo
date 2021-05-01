package bot

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	errLogChan = make(chan upbit.Log)
	eventLogChan = make(chan upbit.Log)
	stdLogChan = make(chan upbit.Log)
	exitLogChan = make(chan upbit.Log)
)

type Bot struct {
	config *upbit.Config
	api    *api.API
}

func New() *Bot {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		logrus.Panic(err)
	}

	return &Bot{config, &api.API{
		Client:          &client.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &client.QuotationClient{Client: &http.Client{}},
	}}
}

func (b *Bot) Run() {
	for coin := range b.config.Coins {
		go b.Tracking(b.config.Coins, coin)
	}

	b.Logging()
}

func (b *Bot) Logging()  {
	exitLogger := upbit.NewLogger("logs/exit.log", logrus.PanicLevel, true)

	errLogger := upbit.NewLogger("logs/error.log", logrus.ErrorLevel, true)
	eventLogger := upbit.NewLogger("logs/log.log", logrus.WarnLevel, false)
	stdLogger := upbit.NewLogger("", logrus.InfoLevel, false)

	for {
		select {
		case errLog := <-errLogChan: errLogger.WithFields(errLog.Fields).WithError(errLog.Msg.(error))
		case eventLog := <-eventLogChan: eventLogger.WithFields(eventLog.Fields).Warn(eventLog.Msg)
		case stdLog := <-stdLogChan: stdLogger.WithFields(stdLog.Fields).Info(stdLog.Msg)
		case exitLog := <-exitLogChan: exitLogger.Panic(exitLog.Msg)
		}
	}
}
