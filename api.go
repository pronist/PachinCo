package upbit

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

// 전일 종가대비 변화율
func GetChangeRate(qc *QuotationClient, market string) float64 {
	day, err := qc.Get("/candles/days", map[string]string{
		"market": market, "count": "1",
	})
	if err != nil {
		log.Panic(err)
	}
	if day, ok := day.([]interface{}); ok {
		// 'day[0]'; Count 를 1로 처리했기 때문
		if candle, ok := day[0].(map[string]interface{}); ok {
			///
			fmt.Printf("\r%s: %.4f%%", market, candle["change_rate"].(float64))
			///
			return candle["change_rate"].(float64)
		}
	}

	return math.NaN()
}

// 현재 자금의 현황
func GetBalances(c *Client) Balances {
	// 가지고 있는 자금의 현황 매핑
	balances := make(Balances)

	// 현재 가지고 있는 자금의 현황을 점검.
	accounts, err := c.Call("GET", "/accounts")
	if err != nil {
		log.Panic(err)
	}
	if accounts, ok := accounts.([]interface{}); ok {
		for idx, _ := range accounts {
			if acc, ok := accounts[idx].(map[string]interface{}); ok {
				f, err := strconv.ParseFloat(acc["balance"].(string), 64)
				if err != nil {
					log.Panic(err)
				}
				balances[acc["currency"].(string)] = f
			}
		}
	}
	///
	fmt.Println(accounts)
	///

	return balances
}

// 매수 평균가
func GetAverageBuyPrice(c *Client, market string) KRW {
	ordersChance, err := c.CallWith("GET", "/orders/chance", Query{"market": market})
	if err != nil {
		log.Panic(err)
	}

	if ordersChance, ok := ordersChance.(map[string]interface{}); ok {
		if acc, ok := ordersChance["bid_account"].(map[string]interface{}); ok {
			f, err := strconv.ParseFloat(acc["avg_buy_price"].(string), 64)
			if err != nil {
				log.Panic(err)
			}

			return KRW(f)
		}
	}

	return KRW(math.NaN())
}

func Order(c *Client, market string, price KRW) string {
	return ""
}

// 주문이 체결될 때까지 기다립니다.
func WaitUntilCompletedOrder(c *Client, done chan int, uuid string)  {
}