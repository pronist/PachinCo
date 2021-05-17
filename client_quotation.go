package upbit

import (
	"github.com/google/go-querystring/query"
	"net/http"
)

type QuotationClient struct {
	*http.Client
}

// Call 은 업비트 API 서버로 요청을 보내되, Quotation API 에 한하여 보내도록 한다.
func (qc *QuotationClient) call(url string, v interface{}) (interface{}, error) {
	values, err := query.Values(v)
	if err != nil {
		panic(err)
	}
	encodedQuery := values.Encode()

	req, err := http.NewRequest("GET", Url+"/"+Version+url+"?"+encodedQuery, nil)
	if err != nil {
		return nil, err
	}

	return getResponse(qc.Client, req)
}
