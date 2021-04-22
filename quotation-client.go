package upbit

import (
	"net/http"
)

var QuotationClient = &http.Client{}

func Get(url string, query Query) (Response, error) {
	encodedQuery := query.Encode()
	req, err := http.NewRequest("GET", Url+"/"+Version+url+"?"+encodedQuery, nil)
	if err != nil {
		return nil, err
	}

	return NewResponse(QuotationClient, req)
}
