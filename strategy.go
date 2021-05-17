package upbit

type Strategy interface {
	prepare(bot *Bot, accounts Accounts)
	run(bot *Bot, accounts Accounts, c *coin, t map[string]interface{}) (bool, error)
}
