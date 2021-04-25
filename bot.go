package upbit

import (
	"log"
)

// 몰빵이 아닌 분산 투자 전략을 위한 것
// 총 자금이 100, 'A' 코인에 최대 10 만큼의 자금 할당시 'A' => 0.1 (비중)
var coins = map[string]float64{
	"BTT": 0.51, // 비트토렌트
	//"AHT", // 아하토큰
	//"MED", // 메디블록
	//"TRX", // 트론
	//"STEEM", // 스팀
	//"EOS", // 이오스
	//"XRP", // 리플
	//"PCI", // 페이코인
	//"ADA", // 에이다
	//"GLM", // 골렘
}

type Bot struct {
	Strategy *Strategy

	// 고루틴에서 보고하기 위한 각종 채널
	Logging chan string
	Err     chan error
}

func NewBot(strategy *Strategy) *Bot {
	return &Bot{strategy, make(chan string), make(chan error)}
}

func (b *Bot) Run() error {
	var err error

	for coin := range coins {
		go b.Strategy.Watch(b.Logging, b.Err, coin)
		go b.Strategy.B(coins, b.Logging, b.Err, coin)
		go b.Strategy.S(coins, b.Logging, b.Err, coin)
	}

	for {
		select {
		case l := <-b.Logging:
			log.Println(l)
		case err = <-b.Err:
			break
		}
	}

	return err
}
