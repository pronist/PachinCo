package upbit

import (
	"github.com/pronist/upbit/api"
	"github.com/sirupsen/logrus"
)

// 몰빵이 아닌 분산 투자 전략을 위한 것
// 총 자금이 100, 'A' 코인에 최대 10 만큼의 자금 할당시 'A' => 0.1 (비중)
var coins = map[string]float64{
	"BTT": 0.2, // 비트토렌트
	"AHT": 0.2, // 아하토큰
	//"MED": 0.1, // 메디블록
	"TRX": 0.2, // 트론
	//"STEEM": 0.1, // 스팀
	"EOS": 0.2, // 이오스
	"XRP": 0.2, // 리플
	//"PCI": 0.1, // 페이코인
	//"ADA": 0.1, // 에이다
	//"GLM": 0.1, // 골렘
}

type Bot struct {
	strategy Strategy
	api      *api.API

	// 고루틴에서 보고하기 위한 각종 채널
	ticker  chan Log
	err     chan Log
	logging chan Log
}

func NewBot(strategy Strategy, api *api.API) *Bot {
	return &Bot{strategy, api, make(chan Log), make(chan Log), make(chan Log)}
}

func (b *Bot) Run() {
	for coin := range coins {
		go b.Watch(coin)
		go b.B(coins, coin)
		go b.S(coins, coin)
	}

	errLogger, err := NewLogger("logs/error.log", logrus.ErrorLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	logLogger, err := NewLogger("logs/log.log", logrus.WarnLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	stdLogger, err := NewLogger("", logrus.InfoLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	for {
		select {
		case t := <-b.ticker:
			stdLogger.WithFields(t.Fields).Info(t.Msg)
		case e := <-b.err:
			if e.Terminate {
				errLogger.WithFields(e.Fields).Fatal(e.Msg)
			} else {
				errLogger.WithFields(e.Fields).Error(e.Msg)
			}
		case l := <-b.logging:
			logLogger.WithFields(l.Fields).Warning(l.Msg)
		}
	}
}
