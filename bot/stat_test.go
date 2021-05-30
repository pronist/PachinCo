package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	stat["KRW-BTC"] = staged
	stat["KRW-BTT"] = staged
	stat["KRW-ADA"] = staged

	stat["KRW-ETH"] = untracked
}

func TestGetMarketsFromStat(t *testing.T) {
	testCases := []struct {
		stat    int
		markets []string
	}{
		{staged, []string{"KRW-ADA", "KRW-BTC", "KRW-BTT"}},
		{untracked, []string{"KRW-ETH"}},
	}

	for _, tc := range testCases {
		for _, m := range getMarketsFromStat(tc.stat) {
			assert.Contains(t, tc.markets, m)
		}
	}
}
