package upbit

import (
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

type claims struct {
	AccessKey    string    `json:"access_key"`
	Nonce        uuid.UUID `json:"nonce"`
	QueryHash    string    `json:"query_hash,omitempty"`
	QueryHashAlg string    `json:"query_hash_alg,omitempty"`
	jwt.StandardClaims
}

type token struct {
	SignedString string
	Type         string
}

// newHS256Token returns *Jwt that is signed with 'secretKey'
func newHS256Token(secretKey string, claims claims) (*token, error) {
	signedString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &token{signedString, "Bearer"}, nil
}
