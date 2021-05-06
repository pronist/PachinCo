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
		LogChan <- Log{Msg: err, Level: logrus.FatalLevel}
	}

	return &Detector{ws: ws, D: make(chan map[string]interface{})}
}

// Search 는 화폐(KRW, BTC, USDT)에 대응하는 마켓에 대해 종목을 검색한다.
// Detector.predicate 조건에 부합하는 종목이 검색되면 Detector.D 채널로 해당 tick 을 내보낸다.
func (d *Detector) Search(currency string) {
	markets, err := upbit.API.GetMarkets()
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.FatalLevel}
	}

	K := funk.Chain(markets).
		Map(func(market map[string]interface{}) string {
			return market["market"].(string)
		}).
		Filter(func(market string) bool {
			return strings.HasPrefix(market, currency)
		})

	for {
		// https://docs.upbit.com/docs/upbit-quotation-websocket
		data := []map[string]interface{}{
			{"ticket": uuid.NewV4()}, // ticket
			{"type": "ticker", "codes": K.Value().([]string), "isOnlySnapshot": true, "isOnlyRealtime": false}, // type
			// format
		}
		if err := d.ws.WriteJSON(data); err != nil {
			LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
		}

		K.ForEach(func(market string) {
			var r map[string]interface{}

			if err = d.ws.ReadJSON(&r); err != nil {
				LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
			}

			f := logrus.Fields{
				"change-rate": r["change_rate"].(float64),
				"price": r["trade_price"].(float64),
			}
			LogChan <- Log{
				Msg: market, Fields: f, Level: logrus.InfoLevel,
			}

			// 이 조건에 충족되면 알림을 보낸다.
			if d.predicate(market, r["trade_price"].(float64)) {
				LogChan <- Log{
					Msg: market, Fields: f, Level: logrus.WarnLevel,
				}

				d.D <- r
			}

			time.Sleep(time.Second * 1)
		})
	}
}

// 트래킹할 종목에 대한 조건이다.
func (d *Detector) predicate(market string, price float64) bool {
	// https://wikidocs.net/21888
	dayCandles, err := upbit.API.GetCandlesDays(market, "2")
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}

	// "변동성 돌파" 한 종목을 트래킹할 조건으로 설정.
	R := dayCandles[1]["high_price"].(float64) - dayCandles[1]["low_price"].(float64)

	return dayCandles[0]["opening_price"].(float64)+(R*upbit.Config.K) < price
}
