package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"time"
)

// Coin 은 전략이 보내는 사인을 받아야하고, 받은 사인을 upbit.Bot 으로 보내야하므로
// Publisher, Observer 두 개의 역할을 다 수행한다.
type Coin struct {
	Name       string                        // 코인의 이름
	Rate       float64                       // 코인에 적용할 비중
	Strategies []Strategy                    // 해당 코인에 사용할 전략
	Limit      float64                       // 코인에 할당된 최대 가격
	Order      float64                       // 한 번 주문시 주문할 가격
	Balance    float64                       // 현재 코인의 밸런스
	Ticker     chan []map[string]interface{} // 틱
}

func NewCoin(name string, rate float64, strategies []Strategy) (*Coin, error) {
	coin := Coin{name, rate, strategies, 0, 0, 0, make(chan []map[string]interface{})}
	coin.Refresh()

	for _, strategy := range strategies {
		strategy.Attach(&coin)
	}

	return &coin, nil
}

func (c *Coin) Update(side string, volume, price float64) {
	c.Try(side, volume, price)
	c.Refresh()
}

// Refresh 메서드는 봇과 업비트와의 계좌 동기화를 위해 정보를 갱신해야 한다.
// 주로 매수/매도를 할 때 정보의 변동이 발생하므로 주문 이후 즉시 처리한다.
func (c *Coin) Refresh() {
	var err error

	Mu.Lock() // 계정 정보를 변경할 떄는 갱신 경쟁을 해서는 안된다.
	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	Accounts, err = upbit.API.NewAccounts()
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
	Mu.Unlock()

	balances, err := upbit.API.GetBalances(Accounts)
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}

	c.Balance = balances[c.Name]

	// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
	total, err := upbit.API.GetTotalBalance(Accounts, balances) // 초기 자금
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}

	// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
	// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
	// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
	c.Limit = total * c.Rate

	// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
	// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
	c.Order = c.Limit * upbit.Config.R

	Logger.
		WithFields(logrus.Fields{
			"total": total, "limit": c.Limit, "order": c.Order, "balance": c.Balance,
		}).
		Log(logrus.InfoLevel, c.Name)
}

// order 메서드는 주문을 하되 Config.Timeout 만큼이 지나가면 주문을 자동으로 취소한다.
// 매수/매도에 둘다 사용한다.
func (c *Coin) Try(side string, volume, price float64) {
	uuid, err := upbit.API.Order("KRW-"+c.Name, side, volume, price)
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
	LogChan <- Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": "KRW-" + c.Name, "volume": volume, "price": price,
		},
		Level: logrus.WarnLevel,
	}

	done := make(chan int)

	timer := time.NewTimer(time.Second * upbit.Config.Timeout)

	go upbit.API.Wait(done, uuid)

	select {
	// 주문이 체결되지 않고 무기한 기다리는 것을 방지하기 위해 타임아웃을 지정한다.
	case <-timer.C:
		err := upbit.API.CancelOrder(uuid)
		if err != nil {
			LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
		}
		LogChan <- Log{
			Msg: "CANCEL",
			Fields: logrus.Fields{
				"coin": c.Name, "side": side, "timeout": time.Second * upbit.Config.Timeout,
			},
			Level: logrus.WarnLevel,
		}
	case <-done:
	}
}
