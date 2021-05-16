package upbit

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
	"time"
)

const (
	SockURL     = "wss://api.upbit.com/websocket"
	SockVersion = "v1"
)

type Detector struct {
	ws *websocket.Conn
	D  chan map[string]interface{}
}

// NewDector 는 새로운 Detector 를 만들고 Detector.ws 에 웹소켓을 설정한다.
func NewDetector() (*Detector, error) {
	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		return nil, err
	}

	return &Detector{ws: ws, D: make(chan map[string]interface{})}, nil
}

// Search 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// Detector.predicate 조건에 부합하는 종목이 검색되면 Detector.D 채널로 해당 tick 을 내보낸다.
func (d *Detector) Run(bot *Bot, currency string, predicate func(b *Bot, t map[string]interface{}) bool) {
	defer func() {
		if err := recover(); err != nil {
			Logger <- Log{Msg: err, Fields: logrus.Fields{"role": "Detector"}, Level: logrus.ErrorLevel}
		}
	}()
	//
	Logger <- Log{Msg: "Detector started...", Level: logrus.DebugLevel}
	//

	markets, err := bot.QuotationClient.Call("/market/all", struct{ isDetail bool }{false})
	if err != nil {
		panic(err)
	}

	// 현재 타겟으로 하고 있는 마켓에 대해서만 디텍팅 해야 한다.
	targetMarkets := funk.Chain(markets.([]map[string]interface{})).
		Map(func(market map[string]interface{}) string { return market["market"].(string) }).
		Filter(func(market string) bool { return strings.HasPrefix(market, currency) }).
		Value().([]string)

	// https://docs.upbit.com/docs/upbit-quotation-websocket
	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": targetMarkets, "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	for {
		if err := d.ws.WriteJSON(data); err != nil {
			panic(err)
		}

		for range targetMarkets {
			var r map[string]interface{}

			if err := d.ws.ReadJSON(&r); err != nil {
				panic(err)
			}

			if predicate(bot, r) {
				d.D <- r
			}

			time.Sleep(time.Millisecond * 300)
		}
		// 마켓마다 개별적으로 고루틴을 만들어서 predicate 를 검증하고는 싶다만,
		// 아쉽게도 predicate 에서 업비트 API 서버에 요청할 일이 있다면 횟수 제한(10)이 걸리므로 문제가 발생한다.
	}
}
