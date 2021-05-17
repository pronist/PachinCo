package upbit

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHS256Token(t *testing.T)  {
	claims := claims{StandardClaims: jwt.StandardClaims{}}

	signedString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(Config.SecretKey))
	assert.NoError(t, err)

	token, err := newHS256Token(Config.SecretKey, claims)
	assert.NoError(t, err)

	assert.Equal(t, "Bearer", token.Type)
	assert.Equal(t, signedString, token.SignedString)
}
