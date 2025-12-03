package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAESGCM_encrypt(t *testing.T) {
	password := "123456"
	data := []byte("hello world")
	var salt [SALT_SIZE]byte
	rand.Read(salt[:])

	encrypted, err := AESGCMEncrypt(password, salt, data)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := AESGCMDecrypt(password, encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, data) {
		t.Fatal("decrypted data is not equal to original data")
	}
}

func TestGenerateBytes(t *testing.T) {
	tSeed := time.Unix(1000, 1000)
	salt := PCG32Bytes(16, tSeed, 0)
	assert.NotNil(t, salt, "should be able to generate salt")
	assert.Equal(t, 16, len(salt), "should be able to generate salt")
}

func TestRandomNum(t *testing.T) {
	tSeed := time.Unix(1000, 0)
	num := RandomNum(tSeed)
	assert.NotNil(t, num, "should be able to generate number")
}
