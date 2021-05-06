package bot

import (
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
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

func NewDetector() *Detector {
	ws, _, err := websocket.DefaultDialer.Dial(SockURL+"/"+SockVersion, nil)
	if err != nil {
		LogChan <- upbit.Log{Msg: err, Level: logrus.FatalLevel}
	}

	return &Detector{ws: ws, D: make(chan map[string]interface{})}
}

// Search 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// Detector.predicate 조건에 부합하는 종목이 검색되면 Detector.D 채널로 해당 tick 을 내보낸다.
func (d *Detector) Run(currency string, predicate func(market string, ticker map[string]interface{}) bool) {
	LogChan <- upbit.Log{Msg: "[DETECTING] START", Level: logrus.InfoLevel}

	markets, err := upbit.API.GetMarkets()
	if err != nil {
		LogChan <- upbit.Log{Msg: err, Level: logrus.FatalLevel}
	}

	K := funk.Chain(markets).
		Map(func(market map[string]interface{}) string {
			return market["market"].(string)
		}).
		Filter(func(market string) bool {
			return strings.HasPrefix(market, currency)
		})

	// https://docs.upbit.com/docs/upbit-quotation-websocket
	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": "ticker", "codes": K.Value().([]string), "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
		// format
	}

	for {
		if err := d.ws.WriteJSON(data); err != nil {
			LogChan <- upbit.Log{Msg: err, Level: logrus.ErrorLevel}
		}

		K.ForEach(func(market string) {
			var r map[string]interface{}

			if err := d.ws.ReadJSON(&r); err != nil {
				LogChan <- upbit.Log{Msg: err, Level: logrus.ErrorLevel}
			}

			if predicate(market, r) {
				d.D <- r
			}

			time.Sleep(time.Millisecond * 100)
		})
		// 마켓마다 개별적으로 go 루틴을 만들어서 predicate 를 검증하고는 싶다만,
		// 아쉽게도 predicate 에서 업비트 API 서버에 요청할 일이 있다면 횟수 제한(10)이 걸리므로 문제가 발생한다.
	}
}
