package observer

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

var(
	errLogger = upbit.NewLogger("logs/error.log", logrus.ErrorLevel, true)
	eventLogger = upbit.NewLogger("logs/log.log", logrus.WarnLevel, false)
	stdLogger = upbit.NewLogger("", logrus.InfoLevel, false)
)

const (
	dbName = "upbit.db"
	marketsBucket = "markets"
)

type Observer struct {
	db *bolt.DB // 여기에는 선정된 종목이 들어간다.

	config *upbit.Config
	api *api.API
}

func New() *Observer {
	config, err := upbit.NewConfig("upbit.config.yml")
	if err != nil {
		errLogger.WithError(err)
	}

	db, err := bolt.Open(dbName, 0666, nil)
	if err != nil {
		errLogger.WithError(err)
	}

	return &Observer{db, config, &api.API{
		Client:          &client.Client{AccessKey: config.KeyPair.AccessKey, SecretKey: config.KeyPair.SecretKey},
		QuotationClient: &client.QuotationClient{Client: &http.Client{}},
	}}
}

// penetration 메서드는 '변동성 돌파' 전략을 구현한다.
func (o *Observer) penetration(daysCandles []map[string]interface{}, price float64) bool {
	K := 0.5

	r := daysCandles[1]["high_price"].(float64) - daysCandles[1]["low_price"].(float64) // 전일 고가, 저가
	openPrice := daysCandles[0]["opening_price"].(float64) // 오늘의 시가

	return openPrice + (r * K) < price
}

// predicate 메서드가 true 값을 반환하면 종목을 추가할 것이다.
func (o *Observer) predicate(market string) (bool, logrus.Fields) {
	price, err := o.api.GetPrice(market)
	if err != nil {
		errLogger.WithError(err)
	}

	daysCandles, err := o.api.GetCandlesDays(market, "6")
	if err != nil {
		errLogger.WithError(err)
	}

	changeRate := fmt.Sprintf(
		"%.2f%%", daysCandles[0]["change_rate"].(float64))

	return o.penetration(daysCandles, price), logrus.Fields{"price": price, "change-rate": changeRate}
}

func (o *Observer) streamGetPrice(callback func(float64))  {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (o *Observer) Run() {
	markets, err := o.api.GetMarkets()
	if err != nil {
		errLogger.WithError(err)
	}

	err = o.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(marketsBucket))
		if err != nil {
			return err
		}

		for {
			for _, market := range markets {
				if market, ok := market["market"].(string); ok && strings.HasPrefix(market, "KRW") {
					if ok, fields := o.predicate(market); ok {
						err := bucket.Put([]byte(market), []byte(""))
						if err != nil {
							return err
						}

						eventLogger.WithFields(fields).Warn(market)
					}

					time.Sleep(time.Millisecond * 100)
				}
			}
		}

		return nil
	})
	if err != nil {
		errLogger.WithError(err)
	}
}