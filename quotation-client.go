package upbit

import (
	"net/http"
)

type QuotationClient struct {
	Client *http.Client
}

func (qc *QuotationClient) Get(url string, query Query) (Response, error) {
	encodedQuery := query.Encode()
	req, err := http.NewRequest("GET", Url+"/"+Version+url+"?"+encodedQuery, nil)
	if err != nil {
		return nil, err
	}

	return NewResponse(qc.Client, req)
}
