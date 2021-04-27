package gateway

import "github.com/dgrijalva/jwt-go"

type request struct {
	Jwt    string
	Method string
	Url    string
	Claims Claims
	Query  Query
}

func NewRequest(secretKey string, method, url string, claims Claims, query Query) (*request, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &request{token, method, url, claims, query}, nil
}
