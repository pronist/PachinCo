package upbit

// Order 테스트용이다.
type T_Order struct {
	Side   string  // 주문의 종류,, 매수/매도
	Volume float64 // 주문 수량
	Price  float64 // 주문 가격
}

var (
	T_Accounts = make(map[string]float64) // 실제로 오더하여 Accounts 를 갱신하는 대신 T_Accounts 에 넣는다.
	T_Orders   = make(map[string]T_Order) // 테스트용 오더 기록을 남긴다.
)
