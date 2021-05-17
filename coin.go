package upbit

import (
	"github.com/sirupsen/logrus"
	"sync"
)

// Coin 은 전략이 보내는 사인을 받아야하고, 받은 사인을 upbit.Bot 으로 보내야하므로
// Publisher, Observer 두 개의 역할을 다 수행한다.
type coin struct {
	name           string                      // 코인의 이름
	rate           float64                     // 코인에 적용할 비중
	onceOrderPrice float64                     // 한 번 주문시 주문할 가격
	limit          float64                     // 코인에 할당된 최대 가격
	t              chan map[string]interface{} // 틱
	// 여러 전략이 주문하여 체결되었을 때 갱신경쟁을 피하기 위한 뮤텍스
	mu sync.Mutex
}

// newCoin 은 새로운 coin 을 만든다. detector 와 strategy 사이의 매개체다.
func newCoin(accounts Accounts, name string, rate float64) (*coin, error) {
	coin := coin{name: name, rate: rate, t: make(chan map[string]interface{})}
	if err := coin.refresh(accounts); err != nil {
		panic(err)
	}

	return &coin, nil
}

// Refresh 메서드는 봇과 업비트와의 계좌 동기화를 위해 정보를 갱신해야 한다.
// 주로 매수/매도를 할 때 정보의 변동이 발생하므로 주문 이후 즉시 처리한다.
func (c *coin) refresh(accounts Accounts) error {
	acc, err := accounts.accounts()
	if err != nil {
		return err
	}

	c.mu.Lock() // 계정 정보를 변경할 떄는 갱신 경쟁을 해서는 안된다.
	defer c.mu.Unlock()

	// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
	totalBalance := getTotalBalance(acc, getBalances(acc)) // 총 매수 자금

	// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
	// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
	// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
	c.limit = totalBalance * c.rate

	// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
	// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
	c.onceOrderPrice = c.limit * Config.R

	//
	logger <- log{
		msg: c.name,
		fields: logrus.Fields{
			"total": totalBalance, "limit": c.limit, "order": c.onceOrderPrice,
		},
		level: logrus.InfoLevel,
	}
	//

	return nil
}
