package gateway

import "net/url"

type Query map[string]string

func (q *Query) Encode() string {
	values := url.Values{}

	for key, value := range *q {
		values.Add(key, value)
	}

	return values.Encode()
}
