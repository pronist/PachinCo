package bot

import (
	"time"

	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
)

const targetMarket = "KRW" // 원화 마켓을 추적한다.

type detector struct {
	d chan map[string]interface{}
}

// newDetector 는 새로운 detector 를 만들고 detector.ws 에 웹소켓을 설정한다.
func newDetector() *detector {
	return &detector{d: make(chan map[string]interface{})}
}

// run 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// predicate 조건에 부합하는 종목이 검색되면 detector.d 채널로 해당 tick 을 내보낸다.
func (d *detector) run(bot *Bot, predicate func(b *Bot, t map[string]interface{}) bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger <- log.Log{Msg: err, Fields: logrus.Fields{"role": "Detector"}, Level: logrus.ErrorLevel}
		}
	}()

	markets, err := bot.QuotationClient.Call("/market/all", struct {
		IsDetail bool `url:"isDetail"`
	}{false})
	if err != nil {
		panic(err)
	}

	targetMarkets, err := getMarketNames(markets.([]map[string]interface{}), targetMarket)
	if err != nil {
		panic(err)
	}

	wsc, err := client.NewWebsocketClient("ticker", targetMarkets, true, false)
	if err != nil {
		panic(err)
	}

	defer wsc.Ws.Close()

	log.Logger <- log.Log{Msg: "Detector started...", Level: logrus.DebugLevel}

	for {
		if err := wsc.Ws.WriteJSON(wsc.Data); err != nil {
			panic(err)
		}

		for range targetMarkets {
			var r map[string]interface{}

			if err := wsc.Ws.ReadJSON(&r); err != nil {
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
