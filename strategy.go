package upbit

import (
	"fmt"
	"github.com/pronist/upbit/api"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

type Strategy struct {
	Api *api.API
}

const (
	F = -0.03 // 보유하지 않은 코인 구입시 고려할 하락 기준, (상승한 것은 구입하지 않음)
	L = -0.03 // 구입 하락 기준
	H = 0.03  // 판매 상승 기준

	// 판매/구입 비중 (첫 코인 구입시도 동일비율 적용)
	// 'A' 코인에 10 만큼 할당이 되었을 때, `R` 이 1.0 이라면 100% 사용하여 주문
	R = 0.49
)

func (s *Strategy) order(logging chan Log, errLog chan Log, coin, side string, volume, price float64) (*api.Accounts, map[string]float64) {
	uuid, err := s.Api.Order("KRW-"+coin, side, volume, price)
	if err != nil {
		errLog <- Log{Msg: err.Error()}
	}
	logging <- Log{
		Msg: fmt.Sprintf("ORDER(`%s`) `%s`", side, "KRW-"+coin),
		Fields: logrus.Fields{
			"volume": volume, "price": price,
		},
	}
	// 주문이 체결 될 때까지 기다린다.
	err = s.Api.WaitUntilCompletedOrder(uuid)
	if err != nil {
		errLog <- Log{Msg: err.Error()}
	}

	// 주문이 체결 된 이후 자금 추적을 위해 변경된 정보를 다시 얻어야 한다.
	accounts, err := s.Api.NewAccounts()
	if err != nil {
		errLog <- Log{Msg: err.Error()}
	}
	balances, err := accounts.GetBalances()
	if err != nil {
		errLog <- Log{Msg: err.Error()}
	}

	return accounts, balances
}

// 봇의 '매수' 전략이 담겨있다.
func (s *Strategy) B(markets map[string]float64, logging chan Log, errLog chan Log, coin string) {
	// 매수 코인 목록에 있어야 한다.
	if r, ok := markets[coin]; ok {
		accounts, err := s.Api.NewAccounts()
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}

		// 계좌가 가지고 있는 총 자산을 구한다. 분할 매수전략을 위해서는 약간의 계산이 필요하다.
		totalBalance, err := accounts.GetTotalBalance(balances) // 초기 자금
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}

		// `maxBalance` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
		// 예를 들어 'KRW-BTT' 의 현재 값이 .1 이므로
		// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
		maxBalance := totalBalance * r

		// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
		// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
		orderBalance := maxBalance * R

		if orderBalance < 5000 {
			errLog <- Log{
				Msg: fmt.Sprintf("Order(B) balance must more than 5000"),
				Fields: logrus.Fields{
					"coin": coin, "max-balance": maxBalance, "order-balance": orderBalance,
				},
				Terminate: true,
			}
		}
		logging <- Log{
			Msg: fmt.Sprintf("Start watching for buying `%s`...", coin),
			Fields: logrus.Fields{
				"total-balance": totalBalance, "max-balance": maxBalance, "order-balance": orderBalance,
			},
		}

		for {
			price, err := s.Api.GetPrice("KRW-" + coin) // 현재 코인 가격
			if err != nil {
				errLog <- Log{Msg: err.Error()}
			}

			volume := math.Floor(orderBalance / price)

			// 현재 코인의 보유 여부
			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수 평균가
				if err != nil {
					errLog <- Log{Msg: err.Error()}
				}

				p := price / avgBuyPrice // 매수평균가 대비 변화율

				// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
				// 5000 KRW 은 업비트의 최소주문 금액이다.
				// 해당 코인의 총 매수금액은 maxBalance 를 벗어나면 안 된다.
				if p-1 <= L && balances["KRW"] > orderBalance && balances["KRW"] >= 5000 && (avgBuyPrice*coinBalance)+orderBalance <= maxBalance {
					accounts, balances = s.order(logging, errLog, coin, "bid", volume, price)
				}
			} else {
				// 해당 코인을 처음 사는 경우
				changeRate, err := s.Api.GetChangeRate("KRW-" + coin) // 전날 대비
				if err != nil {
					errLog <- Log{Msg: err.Error()}
				}
				// 전액 하락률을 기준으로 매수
				if changeRate <= F && balances["KRW"] > orderBalance && balances["KRW"] >= 5000 {
					accounts, balances = s.order(logging, errLog, coin, "bid", volume, price)
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		errLog <- Log{
			Msg: fmt.Sprintf("Not found coin '%s' in supported markets", coin),
		}
	}
}

// 봇의 '매도' 전략이 담겨있다.
func (s *Strategy) S(markets map[string]float64, logging chan Log, errLog chan Log, coin string) {
	if _, ok := markets[coin]; ok {
		logging <- Log{
			Msg: fmt.Sprintf("Start watching for sale `%s`", coin),
		}

		accounts, err := s.Api.NewAccounts()
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}
		balances, err := accounts.GetBalances()
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}

		for {
			// 판매하려면 코인을 가지고 있어야 한다.
			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수평균가
				if err != nil {
					errLog <- Log{Msg: err.Error()}
				}

				price, err := s.Api.GetPrice("KRW-" + coin)
				if err != nil {
					errLog <- Log{Msg: err.Error()}
				}

				p := price / avgBuyPrice
				orderBalance := coinBalance * price

				// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
				if p-1 >= H && orderBalance > 5000 {
					// 전량 매도. (일단 전량매도 전략 실험)
					accounts, balances = s.order(logging, errLog, coin, "ask", coinBalance, price)
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		errLog <- Log{
			Msg: fmt.Sprintf("Not found coin '%s' in supported markets", coin),
		}
	}
}

func (s *Strategy) Watch(logging chan Log, errLog chan Log, coin string) {
	for {
		changeRate, err := s.Api.GetChangeRate("KRW-" + coin)
		if err != nil {
			errLog <- Log{Msg: err.Error()}
		}
		logging <- Log{Msg: coin, Fields: logrus.Fields{
			"change-rate": fmt.Sprintf("%.2f%%", changeRate*100)},
		}
		time.Sleep(1 * time.Second)
	}
	// GetAccounts() 에 대한 요청이 너무 많아 여기서는 생략.
}
