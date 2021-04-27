package gateway

import (
	"net/http"
)

type QuotationClient struct {
	*http.Client
}

func (qc *QuotationClient) Do(url string, query Query) (interface{}, error) {
	encodedQuery := query.Encode()
	req, err := http.NewRequest("GET", Url+"/"+Version+url+"?"+encodedQuery, nil)
	if err != nil {
		return nil, err
	}

	resp, err := NewResponse(qc.Client, req)
	if err != nil {
		return nil, err
	}

	return resp.ByMap()
}
