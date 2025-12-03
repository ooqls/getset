package keydb

import (
	"os"
	"testing"

	"github.com/ooqls/getset/crypto/keys"
	"github.com/stretchr/testify/assert"
)

func initDb(t *testing.T) {
	dbf, err := os.CreateTemp("/tmp", "test-*.db")
	assert.Nilf(t, err, "should not fail to create temp file")

	ca, err := keys.CreateX509CA()
	assert.Nilf(t, err, "should not fail to create CA")

	systemK, err := keys.CreateX509(*ca)
	assert.Nilf(t, err, "should not fail to create x509")

	err = Init(dbf.Name(), *systemK)
	assert.Nilf(t, err, "should not fail to init db")
}

func TestCryptoDatabase_setSystemKey(t *testing.T) {
	initDb(t)

	db := GetCryptoDB()
	systemKey := db.GetSystemKey()
	isSystemk, err := db.IsSystemKey(systemKey)
	assert.Nilf(t, err, "should not fail to check system key")
	assert.Truef(t, isSystemk, "should be system key")

	ca, err := keys.CreateX509CA()
	assert.Nilf(t, err, "should not fail to create CA")

	notSystemKey, err := keys.CreateX509(*ca)
	assert.Nilf(t, err, "should not fail to create x509")

	isSystemk, err = db.IsSystemKey(notSystemKey)
	assert.Nilf(t, err, "should not get an error when checking system key")
	assert.Falsef(t, isSystemk, "should not be system key")
}

func TestCryptoDatabase_InsertKeyPair(t *testing.T) {
	initDb(t)
	db := GetCryptoDB()

	assert.Nilf(t, db.InsertKeyPair("hashedpw", []byte("privateKey"), []byte("publicKey")), "should not fail to insert new key pair")
	privateKey, _, err := db.GetKeyPair("hashedpw")
	assert.Nilf(t, err, "should not fail to get key pair")
	assert.Equalf(t, []byte("privateKey"), privateKey, "should be able to get private key")
}
