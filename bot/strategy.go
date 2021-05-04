package bot

// 옵저버 패턴을 구현하자. Strategy 의 경우 상태의 변화가 있을 경우
// Coin 에게 통지해야 한다.
type Strategy interface {
	Attach(coin *Coin)
	Detach()
	Notify(side string, volume, price float64)
	Run(coin *Coin)
}
