package keys

import "testing"

func TestNewRsaKeyPemBytes(t *testing.T) {
	privateKeyPem, publicKeyPem, err := NewRsaKeyPemBytes()
	if err != nil {
		t.Fatal(err)
	}
	if privateKeyPem == nil {
		t.Fatal("privateKeyPem is nil")
	}
	if publicKeyPem == nil {
		t.Fatal("publicKeyPem is nil")
	}
}

func TestRsaKeyPem(t *testing.T) {
	rsaKey, err := NewRSA()
	if err != nil {
		t.Fatal(err)
	}
	if rsaKey == nil {
		t.Fatal("rsaKey is nil")
	}
	privKeyPem, pubKeyPem := rsaKey.Pem()
	if privKeyPem == nil {
		t.Fatal("privKeyPem is nil")
	}
	if pubKeyPem == nil {
		t.Fatal("pubKeyPem is nil")
	}
}
