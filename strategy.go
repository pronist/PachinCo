package upbit

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math"
	"time"
)

type Strategy struct {
	*Client
	*QuotationClient
	Db *bolt.DB
}

type Balances map[string]float64

const (
	F = -0.05 // 보유하지 않은 코인 구입시 고려할 하락 기준, (상승한 것은 구입하지 않음)
	L = -0.03 // 구입 하락 기준
	H = 0.03  // 판매 상승 기준

	// 판매/구입 비중 (첫 코인 구입시도 동일비율 적용)
	// 'A' 코인에 10 만큼 할당이 되었을 때, `R` 이 1.0 이라면 100% 사용하여 주문
	R = 0.45
)

const (
	dbName         = "bot.db"
	balancesBucket = "balances"
)

func NewStrategy(c *Client, qc *QuotationClient) (*Strategy, error) {
	accounts, err := c.GetAccounts()
	if err != nil {
		return nil, err
	}

	// 현재 자금 현황
	balances, err := c.GetBalances(accounts)
	if err != nil {
		return nil, err
	}
	if len(balances) == 0 {
		return nil, errors.New("balances is empty")
	}

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	strategy := Strategy{c, qc, db}

	// 현재 가지고 있는 코인이 없을 때
	if krw, ok := balances["KRW"]; ok && len(balances) == 1 {
		// 가지고 있는 코인이 없을 때는 초기 자금에 대한 값을 영구적으로 저장할 필요가 있음.
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(balancesBucket))
			if err != nil {
				return err
			}

			log.Printf("[info] started with `%f` KRW...\n", krw)

			// 가장 초창기 자금 (투입된 모든 자금)
			err = bucket.Put([]byte("i"), Float64bytes(krw))
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return &strategy, nil
}

// 봇의 '매수' 전략이 담겨있다.
func (s *Strategy) B(markets map[string]float64, logging chan string, errLog chan error, coin string) {
	// 매수 코인 목록에 있어야 한다.
	if r, ok := markets[coin]; ok {
		err := s.Db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(balancesBucket))

			// 초기 자금 얻기
			i := bucket.Get([]byte("i"))

			// `maxBalance` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
			// 예를 들어 'KRW-BTT' 의 현재 값이 .1 이므로
			// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
			maxBalance := Float64FromBytes(i) * r

			// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
			// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
			orderBalance := maxBalance * R

			logging <- fmt.Sprintf(
				"[info] started watching for buying `%s`, max-balance: %f, order-balance: %f...",
				coin, maxBalance, orderBalance)

			for {
				accounts, err := s.Client.GetAccounts()
				if err != nil {
					return err
				}
				balances, err := s.Client.GetBalances(accounts) // 자금 현황
				if err != nil {
					return err
				}

				price, err := s.QuotationClient.GetPrice("KRW-" + coin) // 현재 코인 가격
				if err != nil {
					return err
				}

				volume := math.Floor(orderBalance / price)

				// 현재 코인의 보유 여부
				if coinBalance, ok := balances[coin]; ok {
					avgBuyPrice, err := s.Client.GetAverageBuyPrice(accounts, coin) // 매수 평균가
					if err != nil {
						return err
					}

					p := price / avgBuyPrice // 매수평균가 대비 변화율

					// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
					// 5000 KRW 은 업비트의 최소주문 금액이다.
					if p-1 <= L && balances["KRW"] > orderBalance && balances["KRW"] >= 5000 && (avgBuyPrice*coinBalance)+orderBalance <= maxBalance {
						uuid, err := s.Client.Order("KRW-"+coin, "bid", volume, price)
						if err != nil {
							return err
						}
						logging <- fmt.Sprintf(
							"[event]―ORDER(B):: %s; order: %f, volume: %f, price: %f",
							"KRW-"+coin, orderBalance, volume, price)

						s.Client.WaitUntilCompletedOrder(errLog, uuid)
					}
				} else {
					// 해당 코인을 처음 사는 경우
					changeRate, err := s.QuotationClient.GetChangeRate("KRW-" + coin) // 전날 대비
					if err != nil {
						return err
					}

					// 전액 하락률을 기준으로 매수
					if changeRate <= F && balances["KRW"] > orderBalance && balances["KRW"] >= 5000 {
						uuid, err := s.Client.Order("KRW-"+coin, "bid", volume, price)
						if err != nil {
							return err
						}
						logging <- fmt.Sprintf(
							"[event]―ORDER(B):: %s; order: %f, volume: %f, price: %f",
							"KRW-"+coin, orderBalance, volume, price)

						s.Client.WaitUntilCompletedOrder(errLog, uuid)
					}
				}

				time.Sleep(1 * time.Second)
			}
		})
		if err != nil {
			errLog <- err
		}
	} else {
		errLog <- fmt.Errorf("not found coin '%s' in supported markets", coin)
	}
}

// 봇의 '매도' 전략이 담겨있다.
func (s *Strategy) S(markets map[string]float64, logging chan string, errLog chan error, coin string) {
	if _, ok := markets[coin]; ok {
		logging <- fmt.Sprintf("[info] started watching for selling `%s`", coin)

		for {
			accounts, err := s.Client.GetAccounts()
			if err != nil {
				errLog <- err
			}

			balances, err := s.Client.GetBalances(accounts)
			if err != nil {
				errLog <- err
			}

			// 판매하려면 코인을 가지고 있어야 한다.
			if coinBalance, ok := balances[coin]; ok {
				avgBuyPrice, err := s.Client.GetAverageBuyPrice(accounts, coin) // 매수평균가
				if err != nil {
					errLog <- err
				}

				price, err := s.GetPrice("KRW-" + coin)
				if err != nil {
					errLog <- err
				}

				p := price / avgBuyPrice

				// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
				if p-1 >= H {
					// 전량 매도. (일단 전량매도 전략 실험)
					uuid, err := s.Client.Order("KRW-"+coin, "ask", coinBalance, price)
					if err != nil {
						errLog <- err
					}
					logging <- fmt.Sprintf(
						"[event]―ORDER(S):: %s; volume: %f, price %f",
						"KRW-"+coin, coinBalance, price)

					s.Client.WaitUntilCompletedOrder(errLog, uuid)
				}
			}

			time.Sleep(1 * time.Second)
		}
	} else {
		errLog <- fmt.Errorf("not found coin '%s' in supported markets", coin)
	}
}

func (s *Strategy) Watch(logging chan string, errLog chan error, coin string) {
	for {
		changeRate, err := s.QuotationClient.GetChangeRate("KRW-" + coin)
		if err != nil {
			errLog <- err
		}
		logging <- fmt.Sprintf("[info]―CR(Y):: %s; %f", "KRW-"+coin, changeRate)

		time.Sleep(1 * time.Second)
	}
	// GetAccounts() 에 대한 요청이 너무 많아 여기서는 생략.
}
