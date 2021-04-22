package upbit

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

const (
	Url     = "https://api.upbit.com"
	Version = "v1"
)

type Client struct {
	AccessKey string
	SecretKey string
}

type Claims struct {
	AccessKey    string    `json:"access_key"`
	Nonce        uuid.UUID `json:"nonce"`
	QueryHash    string    `json:"query_hash,omitempty"`
	QueryHashAlg string    `json:"query_hash_alg,omitempty"`
	jwt.StandardClaims
}

func (c *Client) Do(request *Request) (Response, error) {
	var err error

	client := &http.Client{}

	var req *http.Request
	var body []byte

	if len(request.Query) < 1 {
		body, err = json.Marshal(request.Query)
		if err != nil {
			return nil, err
		}
	}
	req, err = http.NewRequest(request.Method, Url+"/"+Version+request.Url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+request.Jwt)

	return NewResponse(client, req)
}

func (c *Client) Call(method, url string) (Response, error) {
	claims := Claims{
		AccessKey:      c.AccessKey,
		Nonce:          uuid.NewV4(),
		StandardClaims: jwt.StandardClaims{},
	}
	req, err := NewRequest(c.SecretKey, method, url, claims, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) CallWith(method, url string, query Query) (Response, error) {
	encodedQuery := query.Encode()
	hash := sha512.Sum512([]byte(encodedQuery))

	claims := Claims{
		AccessKey:      c.AccessKey,
		Nonce:          uuid.NewV4(),
		QueryHash:      hex.EncodeToString(hash[:]),
		QueryHashAlg:   "SHA512",
		StandardClaims: jwt.StandardClaims{},
	}
	req, err := NewRequest(c.SecretKey, method, url+"?"+encodedQuery, claims, query)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}
