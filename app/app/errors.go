package app

import "fmt"

var (
	ErrRegistryFileNotFound error = fmt.Errorf("registry file not found")
	ErrPrivateKeyNotFound   error = fmt.Errorf("private key not found")
	ErrPublicKeyNotFound    error = fmt.Errorf("public key not found")
)
