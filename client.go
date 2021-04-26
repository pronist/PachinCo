package upbit

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"time"
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

func (c *Client) do(request *Request) (Response, error) {
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

func (c *Client) call(method, url string) (Response, error) {
	claims := Claims{
		AccessKey:      c.AccessKey,
		Nonce:          uuid.NewV4(),
		StandardClaims: jwt.StandardClaims{},
	}
	req, err := NewRequest(c.SecretKey, method, url, claims, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *Client) callWith(method, url string, query Query) (Response, error) {
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

	return c.do(req)
}

// 주문
func (c *Client) Order(market, side string, volume, price float64) (string, error) {
	q := Query{
		"market":   market,
		"side":     side,
		"volume":   fmt.Sprintf("%f", volume),
		"price":    fmt.Sprintf("%f", price),
		"ord_type": "limit",
	}
	order, err := c.callWith("POST", "/orders", q)
	if err != nil {
		return "", err
	}

	if order, ok := order.(map[string]interface{}); ok {
		return order["uuid"].(string), nil
	}

	return "", err
}

// 주문이 체결될 때까지 기다리기.
func (c *Client) waitUntilCompletedOrder(errLog chan Log, uuid string) {
	for {
		order, err := c.callWith("GET", "/order", Query{"uuid": uuid})
		if err != nil {
			errLog <- Log{msg: err.Error()}
		}
		if order, ok := order.(map[string]interface{}); ok {
			if order["state"].(string) == "done" {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// 계좌
func (c *Client) getAccounts() ([]map[string]interface{}, error) {
	var acts []map[string]interface{}

	accounts, err := c.call("GET", "/accounts")
	if err != nil {
		return nil, err
	}
	if accounts, ok := accounts.([]interface{}); ok {
		for idx := range accounts {
			if acc, ok := accounts[idx].(map[string]interface{}); ok {
				acts = append(acts, acc)
			}
		}
	}

	return acts, nil
}

// 현재 자금의 현황
func (c *Client) getBalances(accounts []map[string]interface{}) (Balances, error) {
	// 가지고 있는 자금의 현황 매핑
	balances := make(Balances)

	for _, acc := range accounts {
		balance, err := strconv.ParseFloat(acc["balance"].(string), 64)
		if err != nil {
			return nil, err
		}

		balances[acc["currency"].(string)] = balance
	}

	return balances, nil
}

// 매수 평균가
func (c *Client) getAverageBuyPrice(accounts []map[string]interface{}, coin string) (float64, error) {
	var avgBuyPrice float64

	for _, acc := range accounts {
		if acc["currency"] == coin {
			var err error

			avgBuyPrice, err = strconv.ParseFloat(acc["avg_buy_price"].(string), 64)
			if err != nil {
				return 0, err
			}
			break
		}
	}

	return avgBuyPrice, nil
}
