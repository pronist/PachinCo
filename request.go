package upbit

import "github.com/dgrijalva/jwt-go"

type Request struct {
	Jwt    string
	Method string
	Url    string
	Claims Claims
	Query  Query
}

func NewRequest(secretKey string, method, url string, claims Claims, query Query) (*Request, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &Request{token, method, url, claims, query}, nil
}
