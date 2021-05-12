package bot

type Strategy interface {
	Prepare(accounts Accounts) // 전략 준비
	Run(accounts Accounts, coin *Coin, t map[string]interface{}) (bool, error) // 전략 실행
	Name() string // 전략의 이름
}
