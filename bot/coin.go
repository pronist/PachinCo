package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
)

// Coin 은 전략이 보내는 사인을 받아야하고, 받은 사인을 upbit.Bot 으로 보내야하므로
// Publisher, Observer 두 개의 역할을 다 수행한다.
type Coin struct {
	Name       string     // 코인의 이름
	Rate       float64    // 코인에 적용할 비중
	Strategies []Strategy // 해당 코인에 사용할 전략
	Limit      float64    // 코인에 할당된 최대 가격
	Order      float64    // 한 번 주문시 주문할 가격
	Balance    float64    // 현재 코인의 밸런스
	bot        *Bot       // 코인의 상태를 감지할 봇
}

func NewCoin(name string, rate float64, strategies []Strategy) (*Coin, error) {
	coin := Coin{name, rate, strategies, 0, 0, 0, nil}
	coin.Refresh()

	for _, strategy := range strategies {
		strategy.Attach(&coin)
	}

	return &coin, nil
}

func (c *Coin) Attach(bot *Bot) { c.bot = bot }
func (c *Coin) Detach()         { c.bot = nil }
func (c *Coin) Notify(side string, volume, price float64) {
	c.bot.Update(c, side, volume, price)
}

func (c *Coin) Update(side string, volume, price float64) {
	c.Notify(side, volume, price)
}

// Refresh 메서드는 봇과 업비트와의 계좌 동기화를 위해 정보를 갱신해야 한다.
// 주로 매수/매도를 할 때 정보의 변동이 발생하므로 주문 이후 즉시 처리한다.
func (c *Coin) Refresh() {
	var err error

	Mutex.Lock()
	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	Accounts, err = upbit.API.NewAccounts()
	if err != nil {
		LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
	}
	Mutex.Unlock()

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

	Logger.WithFields(logrus.Fields{
		"total": total, "limit": c.Limit, "order": c.Order, "balance": c.Balance},
	).Log(logrus.InfoLevel, c.Name)
}
