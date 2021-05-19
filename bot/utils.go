package bot

import (
	"strings"

	"github.com/thoas/go-funk"
)

// currency 에 해당하는 마켓 이름만 얻어온다.
func getMarketNames(markets []map[string]interface{}, currency string) []string {
	chain := funk.Chain(markets)

	targetMarkets := chain.Map(func(market map[string]interface{}) string {
		return market["market"].(string)
	}).Filter(func(market string) bool {
		return strings.HasPrefix(market, currency)
	}).Value().([]string)

	return targetMarkets
}
