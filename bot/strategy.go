package bot

type Strategy interface {
	prepare(bot *Bot)
	run(bot *Bot, c *coin, t map[string]interface{}) (bool, error)
}
