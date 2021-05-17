package bot

import (
	"github.com/thoas/go-funk"
	"strings"
)

// 업비트에서 지원하는 모든 마켓 중 currency 에 해당하는 마켓 이름만 얻어온다.
func getMarketNames(bot *Bot, currency string) ([]string, error) {
	markets, err := bot.QuotationClient.Call("/market/all", struct {
		IsDetail bool `url:"isDetail"`
	}{false})
	if err != nil {
		return nil, err
	}

	chain := funk.Chain(markets.([]map[string]interface{}))

	targetMarkets := chain.Map(func(market map[string]interface{}) string {
		return market["market"].(string)
	}).Filter(func(market string) bool {
		return strings.HasPrefix(market, currency)
	}).Value().([]string)

	return targetMarkets, nil
}
