package keys

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("secure")

type SecureValue[T any] interface {
	Delete()
	GetValue() T
}

type JwtSigningKey interface {
	Sign(claims jwt.Claims) (string, *jwt.Token, error)
	Decrypt(token string) (*jwt.Token, error)
	PublicKey() *rsa.PublicKey
}

func ParseJwtKey(privkey, pubkey []byte) (JwtSigningKey, error) {
	rsakey, err := ParseRSA(privkey, pubkey)
	if err != nil {
		return nil, err
	}

	return &JwtKey{
		rsakey: *rsakey,
	}, nil
}

func NewJWTKey(key RSAKey) *JwtKey {
	return &JwtKey{
		rsakey: key,
	}
}

type JwtKey struct {
	rsakey RSAKey
}

func (k *JwtKey) Sign(claims jwt.Claims) (string, *jwt.Token, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, err := token.SignedString(&k.rsakey.privkey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign claim: %v", err)
	}
	return tokenStr, token, nil

}

func (k *JwtKey) Decrypt(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", t.Method)
		}

		return &k.rsakey.pubkey, nil
	})
}

func (k *JwtKey) PublicKey() *rsa.PublicKey {
	return &k.rsakey.pubkey
}
