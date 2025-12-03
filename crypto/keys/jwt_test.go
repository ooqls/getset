package keys

import (
	"encoding/base64"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	rsakey, err := NewRSA()
	assert.Nilf(t, err, "should be able to create rsa key")

	jwtkey := NewJWTKey(*rsakey)
	assert.NotNil(t, jwtkey, "should be able to create jwt key")

	data := []byte("hello world")
	enc, _, err := jwtkey.Sign(jwt.MapClaims{"data": data})
	assert.Nilf(t, err, "should be able to sign")

	dec, err := jwtkey.Decrypt(enc)
	assert.Nilf(t, err, "should be able to verify")

	claims, ok := dec.Claims.(jwt.MapClaims)
	assert.Truef(t, ok, "should be able to verify")
	actualData := claims["data"]
	decodedData, err := base64.StdEncoding.DecodeString(actualData.(string))
	assert.Nilf(t, err, "should be able to decode")

	assert.Equalf(t, data, decodedData, "should be able to verify")
}
