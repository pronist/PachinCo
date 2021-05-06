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
	url     = "wss://api.upbit.com/websocket"
	version = "v1"
)

type Detector struct {
	ws *websocket.Conn
	D  chan map[string]interface{}
}

func NewDetector() *Detector {
	// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
	ws, _, err := websocket.DefaultDialer.Dial(url+"/"+version, nil)
	if err != nil {
		LogChan <- upbit.Log{Msg: err, Level: logrus.FatalLevel}
	}

	return &Detector{ws: ws, D: make(chan map[string]interface{})}
}

// Search 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// Detector.predicate 조건에 부합하는 종목이 검색되면 Detector.D 채널로 해당 tick 을 내보낸다.
func (d *Detector) Search(currency string, predicate func(market string, ticker map[string]interface{}) bool) {
	LogChan <- upbit.Log{
		Msg: "Start searching markets...",
		Fields: logrus.Fields{
			"Currency": currency,
		},
		Level: logrus.InfoLevel,
	}
	
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

		for _, market := range K.Value().([]string) {
			var r map[string]interface{}

			if err := d.ws.ReadJSON(&r); err != nil {
				LogChan <- upbit.Log{Msg: err, Level: logrus.ErrorLevel}
			}

			//LogChan <- upbit.Log{
			//	Msg: market,
			//	Fields: logrus.Fields{
			//		"change-rate": r["change_rate"].(float64),
			//		"price": r["trade_price"].(float64),
			//	},
			//	Level: logrus.InfoLevel,
			//}

			if predicate(market, r) {
				d.D <- r
			}

			time.Sleep(time.Millisecond * 100)
		}
	}
}
