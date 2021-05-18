package bot

import (
	"time"

	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/static"
	"github.com/sirupsen/logrus"
)

// UpbitAccounts 는 실제 업비트 계정을 의미한다.
// 실제로 업비트에 자금을 주문하고 조회하기 때문에 실제 테스트는 어렵다. (업비트엔 샌드박스 모드가 없다.)
type UpbitAccounts struct {
	acc []map[string]interface{}
}

// NewUpbitAccounts 는 업비트로부터 실제 계정을 가져와 새로운 UpbitAccounts 를 반환한다.
func NewUpbitAccounts(b *Bot) (*UpbitAccounts, error) {
	accounts, err := b.Client.Call("GET", "/accounts", nil)
	if err != nil {
		return nil, err
	}
	//
	log.Logger <- log.Log{Msg: "Brought about Accounts From Upbit.", Level: logrus.DebugLevel}
	//

	return &UpbitAccounts{accounts.([]map[string]interface{})}, nil
}

// Order 는 업비트에 주문을 요청한다.
// 주문이 체결되지 않을 경우, Config.Timeout 에 설정된 시간이 지나면 주문을 자동으로 취소한다.
func (acc *UpbitAccounts) order(b *Bot, coin *coin, side string, volume, price float64) (bool, error) {
	a, err := acc.accounts()
	if err != nil {
		return false, err
	}

	balances := getBalances(a)
	var ok bool

	if getAverageBuyPrice(a, coin.name)*balances[coin.name]+coin.onceOrderPrice <= coin.limit && volume*price > minimumOrderPrice {
		done := make(chan int)

		c := targetMarket + "-" + coin.name

		order, err := b.Client.Call("POST", "/orders", struct {
			Market  string  `url:"market"`
			Side    string  `url:"side"`
			Volume  float64 `url:"volume"`
			Price   float64 `url:"price"`
			OrdType string  `url:"ord_type"`
		}{c, side, volume, price, "limit"})

		uuid := order.(map[string]interface{})["uuid"].(string)

		if err != nil {
			log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
		}

		timer := time.NewTimer(time.Second * static.Config.Timeout)

		go acc.wait(b, done, uuid)

		//
		log.Logger <- log.Log{
			Msg: "ORDER",
			Fields: logrus.Fields{
				"coin": coin.name, "side": side, "volume": volume, "price": price,
			},
			Level: logrus.WarnLevel,
		}
		//

		select {
		// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
		case <-timer.C:
			// 주문 취소
			_, err := b.Client.Call("DELETE", "/order", struct {
				Uuid string `url:"uuid"`
			}{uuid})
			if err != nil {
				log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
			}

			log.Logger <- log.Log{
				Msg: "CANCEL",
				Fields: logrus.Fields{
					"coin": coin.name, "side": side, "timeout": time.Second * static.Config.Timeout,
				},
				Level: logrus.WarnLevel,
			}
		case <-done:
			ok = true
		}

		time.Sleep(time.Second * 3) // 주문 바로 갱신하지 않고 잠시동안 기다립니다.

		// 주문 이후 계정 갱신
		accounts, err := b.Client.Call("GET", "/accounts", nil)
		if err != nil {
			return false, err
		}

		acc.acc = accounts.([]map[string]interface{})

		if err := coin.refresh(acc); err != nil {
			return false, err
		}
	}
	return ok, nil
}

// Accounts 는 accounts_utils 에 호환가능한 []map[string]interface{} 형태로 반환한다.
// https://docs.upbit.com/reference#%EC%9E%90%EC%82%B0-%EC%A1%B0%ED%9A%8C
func (acc *UpbitAccounts) accounts() ([]map[string]interface{}, error) {
	return acc.acc, nil
}

// wait 은 주문이 체결되기까지 기다린다.
func (acc *UpbitAccounts) wait(b *Bot, done chan int, uuid string) {
	for {
		order, err := b.Client.Call("GET", "/order", struct {
			Uuid string `url:"uuid"`
		}{uuid})
		if err != nil {
			log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
		}

		if order, ok := order.(map[string]interface{}); ok {
			if state, ok := order["state"].(string); ok && state == "done" {
				done <- 1
			}
		}
		time.Sleep(1 * time.Second)
	}
}
