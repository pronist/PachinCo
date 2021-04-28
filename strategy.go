package upbit

import (
	"fmt"
	"github.com/pronist/upbit/api"
	"github.com/sirupsen/logrus"
	"time"
)

type Strategy interface {
	Buying(args map[string]interface{}) (bool, error)
	BuyingIfNotExistCoin(args map[string]interface{}) (bool, error)
	Sell(args map[string]interface{}) (bool, error)
}

const (
	// 'A' 코인에 10 만큼 할당이 되었을 때, `R` 이 1.0 이라면 100% 사용하여 주문
	R = 0.5
)

func (b *Bot) order(coin, side string, volume, price float64) (*api.Accounts, map[string]float64) {
	uuid, err := b.api.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		b.err <- Log{Msg: err.Error()}
	}
	b.logging <- Log{
		Msg: fmt.Sprintf("ORDER(`%s`) `%s`", side, "KRW-"+coin),
		Fields: logrus.Fields{
			"volume": volume, "price": price,
		},
	}
	// 주문이 체결 될 때까지 기다린다.
	err = b.api.WaitUntilCompletedOrder(uuid)
	if err != nil {
		b.err <- Log{Msg: err.Error()}
	}

	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := b.api.NewAccounts()
	if err != nil {
		b.err <- Log{Msg: err.Error()}
	}
	balances, err := accounts.GetBalances()
	if err != nil {
		b.err <- Log{Msg: err.Error()}
	}

	return accounts, balances
}

// 봇의 '매수' 전략이 담겨있다.
func (b *Bot) B(markets map[string]float64, coin string) {
	// 매수 코인 목록에 있어야 한다.
	if r, ok := markets[coin]; ok {
		accounts, err := b.api.NewAccounts()
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}

		// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
		totalBalance, err := accounts.GetTotalBalance(balances) // 초기 자금
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}

		// `maxBalance` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
		// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
		// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
		maxBalance := totalBalance * r

		// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
		// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
		orderBalance := maxBalance * R

		if orderBalance < 5000 {
			b.err <- Log{
				Msg: fmt.Sprintf("Order(`bid`) balance must more than 5000"),
				Fields: logrus.Fields{
					"coin": coin, "max-balance": maxBalance, "order-balance": orderBalance,
				},
				Terminate: true,
			}
		}
		b.logging <- Log{
			Msg: fmt.Sprintf("Start watching for buying `%s`...", coin),
			Fields: logrus.Fields{
				"total-balance": totalBalance, "max-balance": maxBalance, "order-balance": orderBalance,
			},
		}

		for {
			price, err := b.api.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				b.err <- Log{Msg: err.Error()}
			}

			volume := orderBalance / price

			// 현재 코인의 보유 여부
			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수 평균가
				if err != nil {
					b.err <- Log{Msg: err.Error()}
				}
				if balances["KRW"] > orderBalance && balances["KRW"] >= 5000 && (avgBuyPrice*coinBalance)+orderBalance <= maxBalance {
					// 해당 코인의 총 매수금액은 `maxBalance` 를 벗어나면 안 된다.
					ok, err := b.strategy.Buying(map[string]interface{}{
						"avgBuyPrice": avgBuyPrice, "price": price,
					})
					if err != nil {
						b.err <- Log{Msg: err.Error()}
					}
					if ok {
						accounts, balances = b.order(coin, "bid", volume, price)
					}
				}
			} else {
				if balances["KRW"] > orderBalance && balances["KRW"] >= 5000 {
					ok, err := b.strategy.BuyingIfNotExistCoin(map[string]interface{}{
						"api": b.api, "coin": coin,
					})
					if err != nil {
						b.err <- Log{Msg: err.Error()}
					}
					if ok {
						accounts, balances = b.order(coin, "bid", volume, price)
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		b.err <- Log{
			Msg: fmt.Sprintf("Not found coin '%s' in supported markets", coin),
		}
	}
}

// 봇의 '매도' 전략이 담겨있다.
func (b *Bot) S(markets map[string]float64, coin string) {
	if _, ok := markets[coin]; ok {
		b.logging <- Log{
			Msg: fmt.Sprintf("Start watching for sale `%s`", coin),
		}

		accounts, err := b.api.NewAccounts()
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}

		for {
			// 판매하려면 코인을 가지고 있어야 한다.
			if coinBalance, ok := balances[coin]; ok {
				price, err := b.api.GetPrice("KRW-" + coin)
				if err != nil {
					b.err <- Log{Msg: err.Error()}
				}

				orderBalance := coinBalance * price

				ok, err := b.strategy.Sell(map[string]interface{}{
					"accounts": accounts, "price": price, "coin": coin,
				})
				if err != nil {
					b.err <- Log{Msg: err.Error()}
				}
				// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
				if ok && orderBalance > 5000 {
					// 전량 매도. (일단 전량매도 전략 실험)
					accounts, balances = b.order(coin, "ask", coinBalance, price)
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		b.err <- Log{
			Msg: fmt.Sprintf("Not found coin '%s' in supported markets", coin),
		}
	}
}

func (b *Bot) Watch(coin string) {
	for {
		changeRate, err := b.api.GetChangeRate("KRW-" + coin)
		if err != nil {
			b.err <- Log{Msg: err.Error()}
		}
		b.ticker <- Log{Msg: coin, Fields: logrus.Fields{
			"change-rate": fmt.Sprintf("%.2f%%", changeRate*100)},
		}
		time.Sleep(1 * time.Second)
	}
}
