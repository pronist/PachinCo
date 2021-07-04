package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	states["KRW-BTC"] = staged
	states["KRW-BTT"] = staged
	states["KRW-ADA"] = staged

	states["KRW-ETH"] = untracked
}

func TestGetMarketsFromStates(t *testing.T) {
	testCases := []struct {
		stat    int
		markets []string
	}{
		{staged, []string{"KRW-ADA", "KRW-BTC", "KRW-BTT"}},
		{untracked, []string{"KRW-ETH"}},
	}

	for _, tc := range testCases {
		for _, m := range getMarketsFromStates(tc.stat) {
			assert.Contains(t, tc.markets, m)
		}
	}
}
