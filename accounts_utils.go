package upbit

import (
	"github.com/thoas/go-funk"
	"strconv"
)

// 현재 계좌가 가진 자금 현황을 Map 형태로 바꾼다.
// [{ "currency": "KRW", "balance": ... }, ... ] => {"KRW": balance}
func GetBalances(accounts []map[string]interface{}) map[string]float64 {
	return funk.Reduce(accounts, func(balances map[string]float64, acc map[string]interface{}) map[string]float64 {
		balance, err := strconv.ParseFloat(acc["balance"].(string), 64)
		if err != nil {
			panic(err)
		}
		balances[acc["currency"].(string)] = balance
		return balances
	}, make(map[string]float64)).(map[string]float64)
}

// 해당 코인에 해당 매수 평균가를 얻어온다.
func GetAverageBuyPrice(accounts []map[string]interface{}, coin string) float64 {
	t := funk.Find(accounts, func(acc map[string]interface{}) bool { return acc["currency"].(string) == coin })
	if t, ok := t.(map[string]interface{}); ok {
		avgBuyPrice, err := strconv.ParseFloat(t["avg_buy_price"].(string), 64)
		if err != nil {
			panic(err)
		}
		return avgBuyPrice
	}
	return 0
}

// 해당 계정이 가진 최종 KRW 자금 현황을 얻어온다.
func GetTotalBalance(accounts []map[string]interface{}, balances map[string]float64) float64 {
	return funk.Reduce(funk.Keys(balances), func(totalBalance float64, coin string) float64 {
		totalBalance += GetAverageBuyPrice(accounts, coin) * balances[coin]
		return totalBalance
	}, balances["KRW"]).(float64)
}
