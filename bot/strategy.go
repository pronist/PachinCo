package bot

type Strategy interface {
	Run(coin *Coin) // 전략의 본체
	Prepare()       // 전략 자체를 시작하기 전 준비해야 할 것
}
