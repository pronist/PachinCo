package bot

type Strategy interface {
	Prepare(accounts Accounts)                                                 // 전략 자체를 시작하기 전 준비해야 할 것
	Run(accounts Accounts, coin *Coin, t map[string]interface{}) (bool, error) // 전략의 본체
	Name() string                                                              // 전략의 이름
}
