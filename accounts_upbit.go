package upbit

import (
	"github.com/sirupsen/logrus"
	"time"
)

// UpbitAccounts 는 실제 업비트 계정을 의미한다.
// 실제로 업비트에 자금을 주문하고 조회하기 때문에 실제 테스트는 어렵다. (업비트엔 샌드박스 모드가 없다.)
type UpbitAccounts struct {
	accounts []map[string]interface{}
}

// NewUpbitAccounts 는 업비트로부터 실제 계정을 가져와 새로운 UpbitAccounts 를 반환한다.
func NewUpbitAccounts(b *Bot) (*UpbitAccounts, error) {
	accounts, err := b.Client.Call("GET", "/accounts", nil)
	if err != nil {
		return nil, err
	}
	//
	Logger <- Log{Msg: "Brought about Accounts From Upbit.", Level: logrus.DebugLevel}
	//

	return &UpbitAccounts{accounts.([]map[string]interface{})}, nil
}

// Order 는 업비트에 주문을 요청한다.
// 주문이 체결되지 않을 경우, Config.Timeout 에 설정된 시간이 지나면 주문을 자동으로 취소한다.
func (acc *UpbitAccounts) Order(b *Bot, coin *Coin, side string, volume, price float64) (bool, error) {
	a, err := acc.Accounts()
	if err != nil {
		return false, err
	}

	balances := GetBalances(a)
	var ok bool

	if GetAverageBuyPrice(a, coin.Name)*balances[coin.Name]+coin.OnceOrderPrice <= coin.Limit {
		done := make(chan int)

		c := Market + "-" + coin.Name

		order, err := b.Client.Call("POST", "/orders", struct {
			Market  string  `url:"market"`
			Side    string  `url:"side"`
			Volume  float64 `url:"volume"`
			Price   float64 `url:"price"`
			OrdType string  `url:"ord_type"`
		}{c, side, volume, price, "limit"})

		uuid := order.(map[string]interface{})["uuid"].(string)

		if err != nil {
			Logger <- Log{Msg: err, Level: logrus.ErrorLevel}
		}

		timer := time.NewTimer(time.Second * Config.Timeout)

		go acc.wait(b, done, uuid)

		//
		Logger <- Log{
			Msg: "ORDER",
			Fields: logrus.Fields{
				"coin": coin.Name, "side": side, "volume": volume, "price": price,
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
				Logger <- Log{Msg: err, Level: logrus.ErrorLevel}
			}

			Logger <- Log{
				Msg: "CANCEL",
				Fields: logrus.Fields{
					"coin": coin.Name, "side": side, "timeout": time.Second * Config.Timeout,
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

		acc.accounts = accounts.([]map[string]interface{})

		if err := coin.Refresh(acc); err != nil {
			return false, err
		}
	}
	return ok, nil
}

// Accounts 는 accounts_utils 에 호환가능한 []map[string]interface{} 형태로 반환한다.
// https://docs.upbit.com/reference#%EC%9E%90%EC%82%B0-%EC%A1%B0%ED%9A%8C
func (acc *UpbitAccounts) Accounts() ([]map[string]interface{}, error) {
	return acc.accounts, nil
}

// wait 은 주문이 체결되기까지 기다린다.
func (acc *UpbitAccounts) wait(b *Bot, done chan int, uuid string) {
	for {
		order, err := b.Client.Call("GET", "/order", struct {
			Uuid string `url:"uuid"`
		}{uuid})

		if err != nil {
			Logger <- Log{Msg: err, Level: logrus.ErrorLevel}
		}

		if order, ok := order.(map[string]interface{}); ok {
			if state, ok := order["state"].(string); ok && state == "done" {
				done <- 1
			}
		}
		time.Sleep(1 * time.Second)
	}
}
