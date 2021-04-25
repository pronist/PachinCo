package upbit

import (
	"log"
)

// 몰빵이 아닌 분산 투자 전략을 위한 것
// 총 자금이 100, 'A' 코인에 최대 10 만큼의 자금 할당시 'A' => 0.1 (비중)
var coins = map[string]float64{
	"BTT": 0.2, // 비트토렌트
	"AHT": 0.2, // 아하토큰
	//"MED": 0.1, // 메디블록
	"TRX": 0.2, // 트론
	//"STEEM": 0.1, // 스팀
	"EOS": 0.2, // 이오스
	"XRP": 0.2, // 리플
	//"PCI": 0.1, // 페이코인
	//"ADA": 0.1, // 에이다
	//"GLM": 0.1, // 골렘
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

func (b *Bot) Run() {
	log.Println("[info] bot started...")

	for coin := range coins {
		go b.Strategy.Watch(b.Logging, b.Err, coin)
		go b.Strategy.B(coins, b.Logging, b.Err, coin)
		go b.Strategy.S(coins, b.Logging, b.Err, coin)
	}

	for {
		select {
		case l := <-b.Logging:
			log.Println(l)
		case err := <-b.Err:
			log.Panic(err)
		}
	}
}
