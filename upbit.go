package upbit

import (
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
)

// 업비트의 최소 매도/매수 가격은 '5000 KRW'
const MinimumOrderPrice = 5000

// 계정은 글로벌하게 사용되며 갱신시에는 메모리 동기화를 해야한다.
var Accounts []map[string]interface{}

func init() {
	var err error

	// 계정 상태를 갱신한다. 항상 동기화가 되어있어야 한다.
	// 상태 동기화를 할 때는 보통 주문 직후다.
	Accounts, err = API.NewAccounts()
	if err != nil {
		log.Logger <- log.Log{Msg: err, Level: logrus.PanicLevel}
	}
	//
	log.Logger <- log.Log{Msg: "Brought about Accounts From Upbit.", Level: logrus.InfoLevel}
	//
}
