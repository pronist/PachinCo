package strategy

import (
	"github.com/pronist/upbit/api"
)

type UpDown struct {
	F float64 // 코인을 처음 구매할 떄 고려할 하락 기준
	L float64 // 구입 하락 기준
	H float64 // 판매 상승 기준
}

func (ud *UpDown) Buying(args map[string]interface{}) (bool, error) {
	price, avgBuyPrice := args["price"].(float64), args["avgBuyPrice"].(float64)

	p := price / avgBuyPrice // 매수평균가 대비 변화율

	// 매수평균가보다 현재 코인의 가격의 하락률이 `L` 보다 높은 경우
	return p-1 <= ud.L, nil
}

func (ud *UpDown) BuyingIfNotExistCoin(args map[string]interface{}) (bool, error) {
	api, coin := args["api"].(*api.API), args["coin"].(string)

	changeRate, err := api.GetChangeRate("KRW-" + coin) // 전날 대비
	if err != nil {
		return false, err
	}
	// 전액 하락률을 기준으로 매수
	return changeRate <= ud.F, nil
}

func (ud *UpDown) Sell(args map[string]interface{}) (bool, error) {
	accounts, price, coin := args["accounts"].(*api.Accounts), args["price"].(float64), args["coin"].(string)

	avgBuyPrice, err := accounts.GetAverageBuyPrice(coin) // 매수평균가
	if err != nil {
		return false, err
	}

	p := price / avgBuyPrice

	// 현재 코인의 가격이 '상승률' 만큼보다 더 올라간 경우
	return p-1 >= ud.H, nil
}