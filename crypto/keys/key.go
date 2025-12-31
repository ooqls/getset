package keys

import "crypto/rsa"

type Key interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	PublicKey() (rsa.PublicKey, []byte)
	PrivateKey() (rsa.PrivateKey, []byte)
}
