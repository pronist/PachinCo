package upbit

import "testing"

func TestQuery_Encode(t *testing.T) {
	testCases := []struct {
		query  Query
		expect string
	}{
		{Query{"market": "KRW-BTC"}, "market=KRW-BTC"},
		{Query{"uuids": []string{"9ca023a5-851b-4fec-9f0a-48cd83c2eaae"}}, ""},
	}
	for _, tc := range testCases {
		if r := tc.query.Encode(); r != tc.expect {
			t.Errorf("Query %#v; got %#v, want %#v", tc.query, r, tc.expect)
		}
	}
}
