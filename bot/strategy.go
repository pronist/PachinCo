package bot

type Strategy interface {
	Prepare(accounts Accounts)
	Run(accounts Accounts, coin *Coin, t map[string]interface{}) (bool, error)
	Name() string
}
