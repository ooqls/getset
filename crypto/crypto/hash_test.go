package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgon2HashingAlgorithm_Hash(t *testing.T) {
	var salt [SALT_SIZE]byte
	rand.Read(salt[:])
	algorithm := NewArgon2HashingAlgorithm(64*1024, 3, 2, 32, salt[:])
	hash := algorithm.Hash([]byte("password"))
	assert.NotNil(t, hash, "should be able to hash")
}
