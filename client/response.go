package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Response struct {
	*http.Response
}

func NewResponse(client *http.Client, req *http.Request) (*Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{resp}, nil
}

func (r *Response) ByMap() (interface{}, error) {
	var response interface{}

	text, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(text, &response)
	if err != nil {
		return nil, err
	}

	return response, err
}
