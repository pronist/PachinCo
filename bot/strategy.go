package bot

import (
	"fmt"
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

const (
	MinimumOrderPrice = 5000
)

func (b *Bot) update(coin string, coinRate float64) ([]map[string]interface{}, map[string]float64, float64, float64) {
	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := b.api.NewAccounts()
	if err != nil {
		errLogChan <- upbit.Log{Msg: err}
	}
	balances, err := b.api.GetBalances(accounts)
	if err != nil {
		errLogChan <- upbit.Log{Msg: err}
	}

	// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
	totalBalance, err := b.api.GetTotalBalance(accounts, balances) // 초기 자금
	if err != nil {
		errLogChan <- upbit.Log{Msg: err}
	}

	// `limitOrderPrice` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
	// 예를 들어 'KRW-BTT' 의 비중이 .1 이라면,
	// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
	limitOrderPrice := totalBalance * coinRate

	// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
	// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
	orderBuyingPrice := limitOrderPrice * b.config.R

	eventLogChan <- upbit.Log{
		Msg: "UPDATE",
		Fields: logrus.Fields{
			"coin": coin, "total-balance": totalBalance, "limit-order-price": limitOrderPrice, "order-buying-price": orderBuyingPrice,
		},
	}

	return accounts, balances, limitOrderPrice, orderBuyingPrice
}

func (b *Bot) order(coin, side string, coinRate float64, volume, price float64) ([]map[string]interface{}, map[string]float64, float64, float64) {
	uuid, err := b.api.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		errLogChan <- upbit.Log{Msg: err}
	}
	eventLogChan <- upbit.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"side": side, "market": "KRW-" + coin, "volume": volume, "price": price,
		},
	}

	done := make(chan int)

	timer := time.NewTimer(time.Second * b.config.Timeout)

	go b.api.Wait(done, uuid)

	select {
	case <-timer.C:
		err := b.api.CancelOrder(uuid)
		if err != nil {
			errLogChan <- upbit.Log{Msg: err}
		}
		eventLogChan <- upbit.Log{
			Msg: "CANCEL",
			Fields: logrus.Fields{
				"coin": coin, "side": side, "timeout": time.Second * b.config.Timeout,
			},
		}
		return b.update(coin, coinRate)
	case <-done:
		return b.update(coin, coinRate)
	}
}

func (b *Bot) Tracking(markets map[string]float64, coin string) {
	if r, ok := markets[coin]; ok {
		accounts, balances, limitOrderPrice, orderBuyingPrice := b.update(coin, r)

		stdLogChan <- upbit.Log{
			Msg: "Start watching for Buying and Selling...",
			Fields: logrus.Fields{
				"coin": coin,
			},
		}

		for {
			price, err := b.api.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				errLogChan <- upbit.Log{Msg: err}
			}
			changeRate, err := b.api.GetChangeRate("KRW-" + coin) // 전날 대비
			if err != nil {
				errLogChan <- upbit.Log{Msg: err}
			}

			stdLogChan <- upbit.Log{Msg: coin, Fields: logrus.Fields{
				"change-rate": fmt.Sprintf("%.2f%%", changeRate*100), "price": price,
			}}

			///// 매수 전략
			if balances["KRW"] >= MinimumOrderPrice && balances["KRW"] > orderBuyingPrice && orderBuyingPrice > MinimumOrderPrice {
				volume := orderBuyingPrice / price

				if math.IsInf(volume, 0) {
					exitLogChan <- upbit.Log{Msg: "division by zero"}
				}

				if coinBalance, ok := balances[coin]; ok {
					////// 코인이 있을 때 매수 전략

					// 매수 평균가
					avgBuyPrice, err := b.api.GetAverageBuyPrice(accounts, coin)
					if err != nil {
						errLogChan <- upbit.Log{Msg: err}
					}

					if avgBuyPrice <= 0 {
						exitLogChan <- upbit.Log{Msg: fmt.Sprintf("`avgBuyPrice` must not have value `0`")}
					}

					p := price / avgBuyPrice

					if math.IsInf(p, 0) {
						exitLogChan <- upbit.Log{Msg: "division by zero"}
					}

					// 분할 매수 전략 (하락시 평균단가를 낮추는 전략)

					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
					if p-1 <= b.config.L && avgBuyPrice*coinBalance+orderBuyingPrice <= limitOrderPrice {
						accounts, balances, limitOrderPrice, orderBuyingPrice = b.order(coin, "bid", r, volume, price)
					}
				} else {
					////// 코인을 처음 살 떄의 매수 전략

					orders, err := b.api.GetOrderList("KRW-"+coin, "done")
					if err != nil {
						errLogChan <- upbit.Log{Msg: err}
					}

					// 이 매도에는 시장가 매도가 제외된다. 즉, 웹에서 시장가에 매도한 것이 아니라
					// 봇에서 지정가에 매도한 것만 처리된다.
					askOrders := b.api.GetAskOrders(orders)

					if len(askOrders) == 0 {
						// 해당 코인을 매도한 적이 없는 경우.

						// 전날 하락률을 기준으로 매수
						if changeRate <= b.config.F {
							accounts, balances, limitOrderPrice, orderBuyingPrice = b.order(coin, "bid", r, volume, price)
						}
					} else {
						// 해당 코인을 한 번이라도 매도 한 적 있는 경우.

						latestAskPrice, err := b.api.GetLatestAskPrice(orders)
						if err != nil {
							errLogChan <- upbit.Log{Msg: err}
						}

						if latestAskPrice <= 0 {
							exitLogChan <- upbit.Log{Msg: fmt.Sprintf("`latestAskPrice` must not have value `0`")}
						}

						pp := price / latestAskPrice // 마지막 매도가 대비 변화율

						if math.IsInf(pp, 0) {
							exitLogChan <- upbit.Log{Msg: "division by zero"}
						}

						// 마지막으로 매도한 가격을 기준으로 매수
						if pp-1 <= b.config.F {
							accounts, balances, limitOrderPrice, orderBuyingPrice = b.order(coin, "bid", r, volume, price)
						}
					}
				}
			} else {
				exitLogChan <- upbit.Log{
					Msg: "Minimum order price is `5000 KRW`, insufficient for Buying and Selling",
					Fields: logrus.Fields{
						"coin": coin, "limit-order-price": limitOrderPrice, "order-buying-price": orderBuyingPrice,
					},
				}
			}
			///// 매도 전략
			if coinBalance, ok := balances[coin]; ok {
				orderSellingPrice := coinBalance * price

				// 매수 평균가
				avgBuyPrice, err := b.api.GetAverageBuyPrice(accounts, coin)
				if err != nil {
					errLogChan <- upbit.Log{Msg: err}
				}

				p := price / avgBuyPrice

				if math.IsInf(p, 0) {
					exitLogChan <- upbit.Log{Msg: "division by zero"}
				}

				// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
				if p-1 >= b.config.H && orderSellingPrice > MinimumOrderPrice {
					// 전량 매도 전략
					// 더 높은 기대수익률을 바라보기 어려워짐. 하락 리스크에 조금 더 방어적이지만
					// 너무 수비적이라 조금 더 공격적으로 해도 될 것 같음.

					// 절반 매도 전략 (실험 중)
					// Config.H 만큼 올라갔을 때 절반을 매도, 이후 또 다시 그 만큼 올라가면 매도가능.
					// 전량 매도 전략보다는 더 공격적인 전략
					if orderSellingPrice/2 > MinimumOrderPrice {
						coinBalance = coinBalance / 2
					}

					accounts, balances, limitOrderPrice, orderBuyingPrice = b.order(coin, "ask", r, coinBalance, price)
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		exitLogChan <- upbit.Log{
			Msg: fmt.Errorf("Not found coin '%s' in supported markets", coin),
		}
	}
}
