package upbit

import (
	"net/http"
)

type QuotationClient struct {
	Client *http.Client
}

func (qc *QuotationClient) get(url string, query Query) (Response, error) {
	encodedQuery := query.Encode()
	req, err := http.NewRequest("GET", Url+"/"+Version+url+"?"+encodedQuery, nil)
	if err != nil {
		return nil, err
	}

	return NewResponse(qc.Client, req)
}

// 전일 종가대비 변화율
func (qc *QuotationClient) getChangeRate(market string) (float64, error) {
	day, err := qc.get("/candles/days", map[string]string{
		"market": market, "count": "1",
	})
	if err != nil {
		return 0, err
	}
	if day, ok := day.([]interface{}); ok {
		// 'day[0]'; Count 를 1로 처리했기 때문
		if candle, ok := day[0].(map[string]interface{}); ok {
			return candle["change_rate"].(float64), nil
		}
	}

	return 0, nil
}

// 코인의 현재 가격
func (qc *QuotationClient) getPrice(market string) (float64, error) {
	ticker, err := qc.get("/ticker", Query{"markets": market})
	if err != nil {
		return 0, err
	}
	if ticker, ok := ticker.([]interface{}); ok {
		// 하나의 마켓에 대해서만 처리
		if tick, ok := ticker[0].(map[string]interface{}); ok {
			return tick["trade_price"].(float64), nil
		}
	}

	return 0, err
}
