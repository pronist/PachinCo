package upbit

import (
	"github.com/sirupsen/logrus"
	"time"
)

type detector struct {
	d chan map[string]interface{}
}

// newDetector 는 새로운 detector 를 만들고 detector.ws 에 웹소켓을 설정한다.
func newDetector() (*detector, error) {
	return &detector{d: make(chan map[string]interface{})}, nil
}

func (d *detector) run(bot *Bot, currency string, predicate func(b *Bot, t map[string]interface{}) bool) {
	markets, err := getMarketNames(bot, currency)
	if err != nil {
		panic(err)
	}

	wsc, err := newWebsocketClient("ticker", markets, true, false)

	defer func() {
		if err := recover(); err != nil {
			//
			logger <- log{
				msg: err,
				fields: logrus.Fields{ "role": "Detector" },
				level: logrus.ErrorLevel,
			}
			//
			// 디텍터에 문제가 발생하더라도 로그를 남기고 다시 시작한다.
			d.detect(bot, markets, wsc, predicate)
		}
	}()
	d.detect(bot, markets, wsc, predicate)
}

// detect 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// predicate 조건에 부합하는 종목이 검색되면 detector.d 채널로 해당 tick 을 내보낸다.
func (d *detector) detect(bot *Bot, markets []string, wsc *websocketClient, predicate func(b *Bot, t map[string]interface{}) bool) {
	//
	logger <- log{msg: "Detector started...", level: logrus.DebugLevel}
	//
	for {
		if err := wsc.ws.WriteJSON(wsc.data); err != nil {
			panic(err)
		}

		for range markets {
			var r map[string]interface{}

			if err := wsc.ws.ReadJSON(&r); err != nil {
				panic(err)
			}

			if predicate(bot, r) {
				d.d <- r
			}

			time.Sleep(time.Millisecond * 300)
		}
		// 마켓마다 개별적으로 고루틴을 만들어서 predicate 를 검증하고는 싶다만,
		// 아쉽게도 predicate 에서 업비트 API 서버에 요청할 일이 있다면 횟수 제한(10)이 걸리므로 문제가 발생한다.
	}
}
