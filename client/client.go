package client

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-querystring/query"
	uuid "github.com/satori/go.uuid"
)

const (
	apiURL     = "https://api.upbit.com"
	apiVersion = "v1"
)

type Client struct {
	*http.Client
	AccessKey string
	SecretKey string
}

// Call 은 업비트 API 서버로 요청을 보낸다.
// https://docs.upbit.com/docs/create-authorization-request
//
// 응답은 map[string]interface{}, []map[string]interface{} 중에 하나이므로 받아온 이후에는 변환이 필요하다.
//
// 배열을 전달할 때는 주의해야 한다.
// 구조체를 정의할때 필드에 대해 `url:,numbered" 옵션을 주어야 한다.
func (c *Client) Call(method, url string, v interface{}) (interface{}, error) {
	var body []byte

	claims := claims{
		AccessKey:      c.AccessKey,
		Nonce:          uuid.NewV4(),
		StandardClaims: jwt.StandardClaims{},
	}

	values, err := query.Values(v)
	if err != nil {
		panic(err)
	}
	if len(values) > 0 {
		encodedQuery := values.Encode()

		hash := sha512.Sum512([]byte(encodedQuery))

		claims.QueryHash = hex.EncodeToString(hash[:])
		claims.QueryHashAlg = "SHA512"

		url = url + "?" + encodedQuery

		body, err = json.Marshal(values)
		if err != nil {
			return nil, err
		}
	}

	token, err := newHS256Token(c.SecretKey, claims)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, apiURL+"/"+apiVersion+url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", token.Type+" "+token.SignedString)

	return getResponse(c.Client, req)
}

// getResponse 는 client 에 대해 req 를 실행시키고 map[string]interface{}, 또는 []map[string]interface{} 를 반환한다.
func getResponse(client *http.Client, req *http.Request) (interface{}, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r interface{}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	switch t := r.(type) {
	case []interface{}:
		var a []map[string]interface{}

		for _, item := range t {
			a = append(a, item.(map[string]interface{}))
		}
		r = a
	case map[string]interface{}:
		r = t
	}

	return r, nil
}
