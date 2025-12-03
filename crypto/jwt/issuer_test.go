package jwt

import (
	"testing"

	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/crypto/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIsIssuer(t *testing.T) {
	testutils.InitKeys()
	//	ca, err := keys.CreateX509CA()
	//assert.Nil(t, err)

	//cert, err:= keys.CreateX509(*ca)
	//assert.Nil(t, err)

	subj := "123"
	issuer := NewDefaultJwtTokenIssuer[map[string]string]()
	issuer2 := NewJwtTokenIssuer[map[string]string](&TokenConfiguration{
		Audience: []string{"aud3", "aud4"},
		Issuer:   "issuer2",
	}, keys.JWT())

	tokenStr, _, err := issuer.IssueToken(subj, map[string]string{
		"test": "hello",
	})

	assert.Nil(t, err)

	jwtToken, claims, err := issuer.Decrypt(tokenStr)
	assert.NotNilf(t, jwtToken, "should have gotten a jwt token")
	assert.Nilf(t, err, "should not have gotten an error from is issuer")
	assert.Truef(t, jwtToken.Valid, "token should be valid")
	assert.NotEmpty(t, claims["test"])

	notJwtToken, _, err := issuer2.Decrypt(tokenStr)
	assert.NotNilf(t, err, "should not have gotten an error")
	assert.Falsef(t, notJwtToken.Valid, "token should not be valid")
}
