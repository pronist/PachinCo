package bot

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func (b *Bot) order(coin, side string, volume, price float64) ([]map[string]interface{}, map[string]float64) {
	uuid, err := b.api.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		ErrorLogger.Error(err)
	}
	EventLogger.
		WithFields(logrus.Fields{"volume": volume, "price": price}).
		Warn(fmt.Sprintf("ORDER(`%s`) `%s`", side, "KRW-"+coin))

	// 주문이 체결 될 때까지 기다린다.
	err = b.api.WaitUntilCompletedOrder(uuid)
	if err != nil {
		ErrorLogger.Error(err)
	}

	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := b.api.NewAccounts()
	if err != nil {
		ErrorLogger.Error(err)
	}
	balances, err := b.api.GetBalances(accounts)
	if err != nil {
		ErrorLogger.Error(err)
	}

	return accounts, balances
}

func (b *Bot) B(markets map[string]float64, coin string) {
	if r, ok := markets[coin]; ok {
		////// 이 부분은 고정, 건드릴 필요는 없다.
		accounts, err := b.api.NewAccounts()
		if err != nil {
			ErrorLogger.Error(err)
		}
		balances, err := b.api.GetBalances(accounts)
		if err != nil {
			ErrorLogger.Error(err)
		}

		// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
		totalBalance, err := b.api.GetTotalBalance(accounts, balances) // 초기 자금
		if err != nil {
			ErrorLogger.Error(err)
		}

		// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
		// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
		// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
		limitOrderPrice := totalBalance * r

		// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
		// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
		orderPrice := limitOrderPrice * b.config.R

		// 업비트의 제한으로 최소 매수/매도 주문은 5000 KRW 이상이어야 한다.
		// 그 이하로 주문을 시도하는 것을 에러만을 유발할 뿐이므로 미리 걸러주는 것이 좋다.
		if orderPrice < 5000 {
			errMsg := fmt.Sprintf("Order(`bid`) balance must more than 5000")

			ErrorLogger.
				WithFields(logrus.Fields{"coin": coin, "limit-order-price": limitOrderPrice, "order-price": orderPrice}).
				Error(errMsg)

			exit <- errMsg
		}

		EventLogger.
			WithFields(logrus.Fields{"total-balance": totalBalance, "limit-order-price": limitOrderPrice, "order-price": orderPrice}).
			Warn(fmt.Sprintf("Start watching for buying `%s`...", coin))
		//////

		for {
			price, err := b.api.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				ErrorLogger.Error(err)
			}

			volume := orderPrice / price

			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := b.api.GetAverageBuyPrice(accounts, coin) // 매수 평균가
				if err != nil {
					ErrorLogger.Error(err)
				}
				// 해당 코인의 총 매수금액은 `limitOrderPrice` 를 벗어나면 안 된다.
				if balances["KRW"] > orderPrice && balances["KRW"] >= 5000 && (avgBuyPrice*coinBalance)+orderPrice <= limitOrderPrice {
					////// 코인을 가지고 있을 때의 매수 전략

					p := price / avgBuyPrice // 매수평균가 대비 변화율

					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
					if p-1 <= b.config.L {
						accounts, balances = b.order(coin, "bid", volume, price)
					}

					//////
				}
			} else {
				if balances["KRW"] > orderPrice && balances["KRW"] >= 5000 {
					////// 코인을 처음 살 떄의 매수 전략

					changeRate, err := b.api.GetChangeRate("KRW-" + coin) // 전날 대비
					if err != nil {
						ErrorLogger.Error(err)
					}

					// 전액 하락률을 기준으로 매수
					if changeRate <= b.config.F {
						accounts, balances = b.order(coin, "bid", volume, price)
					}

					//////
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		ErrorLogger.Error(fmt.Sprintf("Not found coin '%s' in supported markets", coin))
	}
}

func (b *Bot) S(markets map[string]float64, coin string) {
	if _, ok := markets[coin]; ok {
		EventLogger.Warn(fmt.Sprintf("Start watching for sale `%s`", coin))

		accounts, err := b.api.NewAccounts()
		if err != nil {
			ErrorLogger.Error(err)
		}
		balances, err := b.api.GetBalances(accounts)
		if err != nil {
			ErrorLogger.Error(err)
		}

		for {
			if coinBalance, ok := balances[coin]; ok {
				price, err := b.api.GetPrice("KRW-" + coin)
				if err != nil {
					ErrorLogger.Error(err)
				}

				orderBalance := coinBalance * price

				if orderBalance > 5000 {
					///// 매도 전략

					avgBuyPrice, err := b.api.GetAverageBuyPrice(accounts, coin) // 매수평균가
					if err != nil {
						ErrorLogger.Error(err)
					}

					p := price / avgBuyPrice

					// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
					if p-1 >= b.config.H {
						// 전량 매도. (일단 전량매도 전략 실험)
						accounts, balances = b.order(coin, "ask", coinBalance, price)
					}

					/////
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		ErrorLogger.Error(fmt.Sprintf("Not found coin '%s' in supported markets", coin))
	}
}
