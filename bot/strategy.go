package bot

type Strategy interface {
	register(bot *Bot) error                                       // 봇이 실행될 때 전략이 최초로 등록될 때
	boot(bot *Bot, c *coin) error                                  // 코인을 생성하고 전략을 실행하기 직전
	run(bot *Bot, c *coin, t map[string]interface{}) (bool, error) // 전략
}
