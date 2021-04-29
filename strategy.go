package upbit

import (
	"fmt"
	"github.com/pronist/upbit/api"
	"github.com/sirupsen/logrus"
	"time"
)

func (b *Bot) order(coin, side string, volume, price float64) (*api.Accounts, map[string]float64) {
	uuid, err := b.api.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		ErrLogger.Error(err)
	}
	LogLogger.
		WithFields(logrus.Fields{"volume": volume, "price": price}).
		Warn(fmt.Sprintf("ORDER(`%s`) `%s`", side, "KRW-"+coin))

	// 주문이 체결 될 때까지 기다린다.
	err = b.api.WaitUntilCompletedOrder(uuid)
	if err != nil {
		ErrLogger.Error(err)
	}

	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := b.api.NewAccounts()
	if err != nil {
		ErrLogger.Error(err)
	}
	balances, err := accounts.GetBalances()
	if err != nil {
		ErrLogger.Error(err)
	}

	return accounts, balances
}

// 봇의 '매수' 전략이 담겨있다.
func (b *Bot) B(markets map[string]float64, coin string) {
	// 매수 코인 목록에 있어야 한다.
	if r, ok := markets[coin]; ok {
		accounts, err := b.api.NewAccounts()
		if err != nil {
			ErrLogger.Error(err)
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			ErrLogger.Error(err)
		}

		// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
		totalBalance, err := accounts.GetTotalBalance(balances) // 초기 자금
		if err != nil {
			ErrLogger.Error(err)
		}

		// `maxBalance` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
		// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
		// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
		maxBalance := totalBalance * r

		// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
		// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
		orderBalance := maxBalance * b.config.R

		if orderBalance < 5000 {
			errMsg := fmt.Sprintf("Order(`bid`) balance must more than 5000")

			ErrLogger.
				WithFields(logrus.Fields{"coin": coin, "max-balance": maxBalance, "order-balance": orderBalance}).
				Error(errMsg)

			Exit <- errMsg
		}
		LogLogger.
			WithFields(logrus.Fields{"total-balance": totalBalance, "max-balance": maxBalance, "order-balance": orderBalance}).
			Warn(fmt.Sprintf("Start watching for buying `%s`...", coin))

		for {
			price, err := b.api.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				ErrLogger.Error(err)
			}

			volume := orderBalance / price

			// 현재 코인의 보유 여부
			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수 평균가
				if err != nil {
					ErrLogger.Error(err)
				}
				if balances["KRW"] > orderBalance && balances["KRW"] >= 5000 && (avgBuyPrice*coinBalance)+orderBalance <= maxBalance {
					// 해당 코인의 총 매수금액은 `maxBalance` 를 벗어나면 안 된다.
					p := price / avgBuyPrice // 매수평균가 대비 변화율

					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
					if p-1 <= b.config.L {
						accounts, balances = b.order(coin, "bid", volume, price)
					}
				}
			} else {
				if balances["KRW"] > orderBalance && balances["KRW"] >= 5000 {
					changeRate, err := b.api.GetChangeRate("KRW-" + coin) // 전날 대비
					if err != nil {
						ErrLogger.Error(err)
					}

					// 전액 하락률을 기준으로 매수
					if changeRate <= b.config.F {
						accounts, balances = b.order(coin, "bid", volume, price)
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		ErrLogger.Error(fmt.Sprintf("Not found coin '%s' in supported markets", coin))
	}
}

// 봇의 '매도' 전략이 담겨있다.
func (b *Bot) S(markets map[string]float64, coin string) {
	if _, ok := markets[coin]; ok {
		LogLogger.Warn(fmt.Sprintf("Start watching for sale `%s`", coin))

		accounts, err := b.api.NewAccounts()
		if err != nil {
			ErrLogger.Error(err)
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			ErrLogger.Error(err)
		}

		for {
			// 판매하려면 코인을 가지고 있어야 한다.
			if coinBalance, ok := balances[coin]; ok {
				price, err := b.api.GetPrice("KRW-" + coin)
				if err != nil {
					ErrLogger.Error(err)
				}

				orderBalance := coinBalance * price

				avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수평균가
				if err != nil {
					ErrLogger.Error(err)
				}

				p := price / avgBuyPrice

				// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
				if p-1 >= b.config.H && orderBalance > 5000 {
					// 전량 매도. (일단 전량매도 전략 실험)
					accounts, balances = b.order(coin, "ask", coinBalance, price)
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		ErrLogger.Error(fmt.Sprintf("Not found coin '%s' in supported markets", coin))
	}
}

func (b *Bot) Watch(coin string) {
	for {
		changeRate, err := b.api.GetChangeRate("KRW-" + coin)
		if err != nil {
			ErrLogger.Error(err)
		}

		StdLogger.
			WithFields(logrus.Fields{"change-rate": fmt.Sprintf("%.2f%%", changeRate*100)}).
			Info(coin)

		time.Sleep(1 * time.Second)
	}
}
