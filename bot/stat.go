package bot

import "github.com/thoas/go-funk"

// 마켓의 추적 상태를 나타내는 상수다.
// 추적된 마켓에 대해서는 tracked 상수가, 해제된 마켓에 대해서는 untracked 가 사용된다.
const (
	tracked = iota
	untracked
)

// tracked 상태인 마켓에 대해서는 전략 및 틱이 실행될 것이며
// 그렇지 않은 마켓에 대해서는 중지 될 것이다.
var stat = make(map[string]int)

// stat 에서 마켓 상태에 따른 키값들을 얻어온다.
func getMarketsFromStat(marketState int) []string {
	return funk.Filter(funk.Keys(stat), func(market string) bool {
		return stat[market] == marketState
	}).([]string)
}
