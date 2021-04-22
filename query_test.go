package upbit

import "testing"

func TestQuery_Encode(t *testing.T) {
	testCases := []struct {
		query  Query
		expect string
	}{
		{Query{"market": "KRW-BTC"}, "market=KRW-BTC"},
	}
	for _, tc := range testCases {
		if r := tc.query.Encode(); r != tc.expect {
			t.Errorf("Query %#v; got %#v, want %#v", tc.query, r, tc.expect)
		}
	}
}
