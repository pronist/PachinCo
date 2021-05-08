package bot

import (
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/log"
	"github.com/sirupsen/logrus"
	"sync"
)

// Coin 은 전략이 보내는 사인을 받아야하고, 받은 사인을 upbit.Bot 으로 보내야하므로
// Publisher, Observer 두 개의 역할을 다 수행한다.
type Coin struct {
	Name           string                      // 코인의 이름
	Rate           float64                     // 코인에 적용할 비중
	OnceOrderPrice float64                     // 한 번 주문시 주문할 가격
	Limit          float64                     // 코인에 할당된 최대 가격
	T              chan map[string]interface{} // 틱
	// 여러 전략이 주문하여 체결되었을 때 갱신경쟁을 피하기 위한 뮤텍스
	mu sync.Mutex
}

func NewCoin(name string, rate float64) (*Coin, error) {
	coin := Coin{Name: name, Rate: rate, T: make(chan map[string]interface{})}
	if err := coin.Refresh(); err != nil {
		panic(err)
	}

	return &coin, nil
}

// order 메서드는 주문을 하되 Config.Timeout 만큼이 지나가면 주문을 자동으로 취소한다.
// 매수/매도에 둘다 사용한다.
func (c *Coin) Order(side string, volume, price float64) {
	log.Logger <- log.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": "KRW-" + c.Name, "volume": volume, "price": price,
		},
		Level: logrus.WarnLevel,
	}
	//
	//done := make(chan int)
	//
	//uuid, err := upbit.API.Order("KRW-"+c.Name, side, fmt.Sprintf("%f", volume), fmt.Sprintf("%f", price), "limit")
	//if err != nil {
	//	LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	//}
	//LogChan <- Log{
	//	Msg: "ORDER",
	//	Fields: logrus.Fields{
	//		"side": side, "market": "KRW-" + c.Name, "volume": volume, "price": price,
	//	},
	//	Level: logrus.WarnLevel,
	//}
	//
	//timer := time.NewTimer(time.Second * upbit.Config.Timeout)
	//
	//go upbit.API.Wait(done, uuid)
	//
	//select {
	//// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
	//case <-timer.C:
	//	err := upbit.API.CancelOrder(uuid)
	//	if err != nil {
	//		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	//	}
	//	LogChan <- Log{
	//		Msg: "CANCEL",
	//		Fields: logrus.Fields{
	//			"coin": c.Name, "side": side, "timeout": time.Second * upbit.Config.Timeout,
	//		},
	//		Level: logrus.WarnLevel,
	//	}
	//case <-done:
	//}
	//
	//time.Sleep(time.Second * 3) // 주문 바로 갱신하지 않고 잠시동안 기다립니다.
	//
	//if err := c.Refresh(); err != nil {
	//	LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	//}
}

// Refresh 메서드는 봇과 업비트와의 계좌 동기화를 위해 정보를 갱신해야 한다.
// 주로 매수/매도를 할 때 정보의 변동이 발생하므로 주문 이후 즉시 처리한다.
func (c *Coin) Refresh() error {
	var err error

	c.mu.Lock() // 계정 정보를 변경할 떄는 갱신 경쟁을 해서는 안된다.
	defer c.mu.Unlock()

	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	upbit.Accounts, err = upbit.API.NewAccounts()
	if err != nil {
		return err
	}

	balances, err := upbit.API.GetBalances(upbit.Accounts)
	if err != nil {
		return err
	}

	// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
	totalBalance, err := upbit.API.GetTotalBalance(upbit.Accounts, balances) // 초기 자금
	if err != nil {
		return err
	}

	// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
	// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
	// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
	c.Limit = totalBalance * c.Rate

	// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
	// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
	c.OnceOrderPrice = c.Limit * upbit.Config.R

	return nil
}
