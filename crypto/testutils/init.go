package testutils

import (
	"fmt"

	"github.com/ooqls/getset/crypto/keys"
)

func InitKeys() {
	privKey, pubKey, err := keys.NewRsaKeyPemBytes()
	if err != nil {
		panic(fmt.Errorf("failed to generate new RSA key: %v", err))
	}

	err = keys.InitJwt(privKey, pubKey)
	if err != nil {
		panic(fmt.Errorf("failed to init keys: %v", err))
	}

	err = keys.InitRSA(privKey, pubKey)
	if err != nil {
		panic(fmt.Errorf("failed to init keys: %v", err))
	}
}
