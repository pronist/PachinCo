package upbit

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Response interface{}

func NewResponse(client *http.Client, req *http.Request) (Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Get response text from *http.Response.Body
	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response

	// Get mapped response
	switch rune(text[0]) {
	case '[': response = make([]interface{}, 0)
	case '{': response = make(map[string]interface{})
	}

	err = json.Unmarshal(text, &response)
	if err != nil {
		return nil, err
	}

	return response, err
}