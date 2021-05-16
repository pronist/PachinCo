package upbit

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var qc = &QuotationClient{http.DefaultClient}

func TestQuotationClient_Call(t *testing.T) {
	markets, err := qc.Call(
		"/market/all",
		struct {
			IsDetails bool `url:"isDetails"`
		}{false})
	assert.NoError(t, err)

	assert.Contains(t, markets, map[string]interface{}{
		"market": "KRW-BTC", "korean_name": "비트코인", "english_name": "Bitcoin",
	})
}
