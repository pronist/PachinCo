package bot

const minimumOrderPrice = 5000 // 업비트의 최소 매도/매수 가격은 '5000 KRW'

const (
	b = "bid" // 매수
	s = "ask" // 매도
)

type Accounts interface {
	// order 메서드는 주문을 하되 Config.Timeout 만큼이 지나가면 주문을 자동으로 취소한다.
	// 매수/매도에 둘다 사용한다.
	order(b *Bot, c *coin, side string, volume, price float64) (bool, error)
	// 내부에 있는 upbit.API 에서의 접근을 위해 accounts 를 반환해야 한다.
	accounts() ([]map[string]interface{}, error)
}
