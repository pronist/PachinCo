package client

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/pronist/upbit/static"
	"github.com/stretchr/testify/assert"
)

func TestNewHS256Token(t *testing.T) {
	claims := claims{StandardClaims: jwt.StandardClaims{}}

	signedString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(static.Config.SecretKey))
	assert.NoError(t, err)

	token, err := newHS256Token(static.Config.SecretKey, claims)
	assert.NoError(t, err)

	assert.Equal(t, "Bearer", token.Type)
	assert.Equal(t, signedString, token.SignedString)
}
