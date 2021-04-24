package upbit

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
)

type Bot struct {
	*Client
	*QuotationClient
	Db *bolt.DB
}

const (
	F = -0.05 // 보유하지 않은 코인 구입시 고려할 하락 기준, (상승한 것은 구입하지 않음)
	L = -0.03 // 구입 하락 기준
	H = 0.03 // 판매 상승 기준
	R = 0.2 // 판매/구입 비중 (첫 코인 구입시도 동일비율 적용)
)

const (
	dbName = "bot.db"
	balancesBucket = "balances"
)

// KRW 단위를 의미. 코인이 가지는 단위랑은 분명한 분리를 위해 선언되었음.
type KRW float64

var markets = []string{
	"KRW-BTT", // 비트토렌트
	//"KRW-AHT", // 아하토큰
	//"KRW-MED", // 메디블록
	//"KRW-TRX", // 트론
	//"KRW-STEEM", // 스팀
	//"KRW-EOS", // 이오스
	//"KRW-XRP", // 리플
	//"KRW-PCI", // 페이코인
	//"KRW-ADA", // 에이다
	//"KRW-GLM", // 골렘
}

// 몰빵이 아닌 분산 투자 전략을 위한 것
// 총 자금이 100, 'A' 코인에 최대 10 만큼의 자금 할당시 'A' => 0.1 (비중)markets
// 코인당 할당된 보유 금액 단위는 'KRW'
// Balances 의 총 합은 처음 가지고 있던 자금의 KRW 와 같아야 함
var rates = map[string]float64{
	"BTT" : 0.1,
	//"AHT",
	//"MED",
	//"TRX",
	//"STEEM",
	//"EOS",
	//"XRP",
	//"PCI",
	//"ADA",
	//"GLM",
}

func NewBot(c *Client, qc *QuotationClient) (*Bot, error) {
	// 현재 자금 현황
	balances := GetBalances(c)
	if len(balances) == 0 {
		return nil, errors.New("balances is empty")
	}

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	bot := Bot{c, qc, db}

	// 가지고 있는 모든 자금 저장
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(balancesBucket))
		if err != nil {
			return err
		}

		bot.save(balances, bucket)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 현재 가지고 있는 코인이 없을 때
	if krw, ok := balances["KRW"]; ok && len(balances) == 1 {
		// 가지고 있는 코인이 없을 때는 초기 자금에 대한 값을 영구적으로 저장할 필요가 있음.
		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(balancesBucket))

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

	return &bot, nil
}

func (b *Bot) order(bucket *bolt.Bucket, balances Balances, coin string, orderPrice KRW) error {
	// 주문 접수 후 체결까지 기다린 후에 자금상황을 갱신해야 합니다.
	done := make(chan int)

	go WaitUntilCompletedOrder(b.Client, done, Order(b.Client, "KRW-" + coin, orderPrice))
	<-done

	// 현재 자금 현황 저장
	// 자금 갱신의 경우 주문접수 후, 체결이 된 이후에 해야 함.
	return b.save(balances, bucket)
}

func (b *Bot) save(balances Balances, bucket *bolt.Bucket) error {
	encodedBalances, err := balances.Encode()
	if err != nil {
		return err
	}

	return bucket.Put([]byte("b"), encodedBalances)
}

// 봇의 매수전략이 담겨있습니다.
func (b *Bot) Buy(coin string) error {
	var err error

	if r, ok := rates[coin]; ok {
		err = b.Db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(balancesBucket))

			// 초기 자금 얻기
			i := bucket.Get([]byte("i"))

			// `maxBalance` 는 분배된 비율에 따라 초기자금 대비 최대로 구입 가능한 비율이다.
			// 예를 들어 'KRW-BTT' 의 현재 값이 .1 이므로
			// 초기자금 대비 10% 에 대해서만 구입하도록 하기 위한 것이다.
			maxBalance := KRW(Float64FromBytes(i) * r)

			// 한번에 주문할 수 있는 가격, `maxBalance` 에서 `R` 만큼만 주문한다.
			// 총 자금이 100, `maxBalance` 가 10인 경우 `R` 이 .2 이므로 10의 20% 에 해당하는 2 만큼만 주문
			orderPrice := maxBalance * R

			balances, err := DecodeBalances(bucket.Get([]byte("b")))
			if err != nil {
				return err
			}

			// 현재 코인의 보유 여부
			if balance, ok := balances[coin]; ok {
				// 구매하려는 코인을 이미 가지고 있는 경우
				// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우

				//avgBuyPrice := GetAverageBuyPrice(b.Client, coin) // 매수 평균가
				//
				//if avgBuyPrice <= L && (KRW(balance) + orderPrice) < maxBalance {
				//	//
				//	fmt.Printf("L; maxBalance: %.1f, Balance: %.1f, Order: %.1f", maxBalance, balance, orderPrice)
				//	//
				//	return b.order(bucket, balances, coin, orderPrice)
				//}
			} else {
				// 해당 마켓에 대한 코인을 처음 구매하는 경우
				// 전액 하락률을 기준으로 매수
				changeRate := GetChangeRate(b.QuotationClient, "KRW-" + coin) // 전날 대비

				if changeRate <= F && (KRW(balance) + orderPrice) < maxBalance {
					//
					fmt.Printf("F; maxBalance: %.1f, Balance: %.1f, Order: %.1f", maxBalance, balance, orderPrice)
					//
					return b.order(bucket, balances, coin, orderPrice)
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not found coin '%s' in supported markets", coin)
	}

	return err
}