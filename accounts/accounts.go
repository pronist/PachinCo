package accounts

import (
	"fmt"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
	"time"
)

type Accounts struct {
	accounts []map[string]interface{}
}

func New() (*Accounts, error) {
	accounts, err := upbit.API.NewAccounts()
	if err != nil {
		return nil, err
	}
	//
	log.Logger <- log.Log{Msg: "Brought about Accounts From Upbit.", Level: logrus.InfoLevel}
	//

	return &Accounts{accounts}, nil
}

func (acc *Accounts) Order(coin *bot.Coin, side string, volume, price float64, t map[string]interface{}) (bool, error) {
	done := make(chan int)
	var ok bool

	c := upbit.Market + "-" + coin.Name

	uuid, err := upbit.API.Order(c, side, fmt.Sprintf("%f", volume), fmt.Sprintf("%f", price), "limit")
	if err != nil {
		log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
	}

	timer := time.NewTimer(time.Second * upbit.Config.Timeout)

	go upbit.API.Wait(done, uuid)

	//
	log.Logger <- log.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"coin": coin.Name, "side": side, "volume": volume, "price": price, "change-rate": t["change_rate"].(float64),
		},
		Level: logrus.WarnLevel,
	}
	//

	select {
	// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
	case <-timer.C:
		err := upbit.API.CancelOrder(uuid)
		if err != nil {
			log.Logger <- log.Log{Msg: err, Level: logrus.ErrorLevel}
		}

		log.Logger <- log.Log{
			Msg: "CANCEL",
			Fields: logrus.Fields{
				"coin": coin.Name, "side": side, "timeout": time.Second * upbit.Config.Timeout,
			},
			Level: logrus.WarnLevel,
		}
	case <-done:
		ok = true
	}

	time.Sleep(time.Second * 3) // 주문 바로 갱신하지 않고 잠시동안 기다립니다.

	// 주문 이후 계정 갱신
	accounts, err := upbit.API.NewAccounts()
	if err != nil {
		return false, err
	}

	acc.accounts = accounts

	if err := coin.Refresh(acc); err != nil {
		return false, err
	}

	return ok, nil
}

func (acc *Accounts) Accounts() []map[string]interface{} {
	return acc.accounts
}
