package main

import (
	"github.com/pronist/upbit"
	"net/http"
	"sync"
)

const (
	accessKey = "UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4"
	secretKey = "2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn"
)

var markets = []string{
	"KRW-BTT", // 비트토렌트
	"KRW-AHT", // 아하토큰
	"KRW-MED", // 메디블록
	"KRW-TRX", // 트론
	"KRW-STEEM", // 스팀
	"KRW-EOS", // 이오스
	"KRW-XRP", // 리플
	"KRW-PCI", // 페이코인
	"KRW-ADA", // 에이다
	"KRW-GLM", // 골렘
}

func main() {
	var wg sync.WaitGroup

	bot := upbit.Bot{
		Client: &upbit.Client{AccessKey: accessKey, SecretKey: secretKey},
		QuotationClient: &upbit.QuotationClient{Client: &http.Client{}},
	}
	for _, market := range markets {
		wg.Add(1)
		go bot.Watch(market)
	}

	wg.Wait()
}
