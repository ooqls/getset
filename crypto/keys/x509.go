package keys

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"time"
)

type option func(*x509.Certificate)

func WithCommonName(name string) option {
	return func(c *x509.Certificate) {
		c.Subject.CommonName = name
	}
}

func WithDNSNames(dnsNames []string) option {
	return func(c *x509.Certificate) {
		c.DNSNames = dnsNames
	}
}

func WithIPAddresses(ipAddresses []net.IP) option {
	return func(c *x509.Certificate) {
		c.IPAddresses = ipAddresses
	}
}
func WithEmailAddresses(emailAddresses []string) option {
	return func(c *x509.Certificate) {
		c.EmailAddresses = emailAddresses
	}
}

func WithPermittedDNSDomainsCritical(critical bool) option {
	return func(c *x509.Certificate) {
		c.PermittedDNSDomainsCritical = critical
	}
}

func WithPermittedDNSDomains(dnsNames []string) option {
	return func(c *x509.Certificate) {
		c.PermittedDNSDomains = dnsNames
	}
}
func WithPermittedEmailAddresses(emailAddresses []string) option {
	return func(c *x509.Certificate) {
		c.PermittedEmailAddresses = emailAddresses
	}
}

func WithPermittedURIDomains(uriDomains []string) option {
	return func(c *x509.Certificate) {
		c.PermittedURIDomains = uriDomains
	}
}

func WithNotBefore(t time.Time) option {
	return func(c *x509.Certificate) {
		c.NotBefore = t
	}
}
func WithNotAfter(t time.Time) option {
	return func(c *x509.Certificate) {
		c.NotAfter = t
	}
}
func WithKeyUsage(ku x509.KeyUsage) option {
	return func(c *x509.Certificate) {
		c.KeyUsage = ku
	}
}
func WithExtKeyUsage(eku []x509.ExtKeyUsage) option {
	return func(c *x509.Certificate) {
		c.ExtKeyUsage = eku
	}
}

func WithTemplate(template x509.Certificate) option {
	return func(c *x509.Certificate) {
		*c = template
	}
}

func CreateX509CACertificate(opts ...option) (*x509.Certificate, *rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())), 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	pkBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key: %v", err)
	}

	hash := sha1.Sum(pkBytes)

	template := &x509.Certificate{
		SerialNumber:                big.NewInt(time.Now().UnixNano()),
		NotBefore:                   time.Now(),
		NotAfter:                    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:                    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:                 []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		Subject:                     pkix.Name{CommonName: ""},
		IsCA:                        true,
		Issuer:                      pkix.Name{CommonName: ""},
		PublicKey:                   &privKey.PublicKey,
		SignatureAlgorithm:          x509.SHA256WithRSA,
		BasicConstraintsValid:       true,
		SubjectKeyId:                hash[:],
		AuthorityKeyId:              hash[:],
		CRLDistributionPoints:       []string{},
		OCSPServer:                  []string{},
		EmailAddresses:              []string{},
		DNSNames:                    []string{},
		IPAddresses:                 []net.IP{},
		PermittedDNSDomainsCritical: false,
		PermittedDNSDomains:         []string{},
		PermittedEmailAddresses:     []string{},
		PermittedURIDomains:         []string{},
		PublicKeyAlgorithm:          x509.RSA,
	}

	for _, o := range opts {
		o(template)
	}

	certDER, err := x509.CreateCertificate(rand.New(rand.NewSource(time.Now().UnixNano())), template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, privKey, nil
}

func CreateX509Certificate(ca X509, opts ...option) (*x509.Certificate, *rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())), 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:                big.NewInt(time.Now().UnixNano()),
		NotBefore:                   ca.crt.NotBefore,
		NotAfter:                    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:                    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:                 []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		Subject:                     ca.crt.Subject,
		Issuer:                      ca.crt.Subject,
		PublicKey:                   &ca.crt.PublicKey,
		SignatureAlgorithm:          x509.SHA256WithRSA,
		BasicConstraintsValid:       true,
		SubjectKeyId:                ca.crt.SubjectKeyId,
		AuthorityKeyId:              ca.crt.AuthorityKeyId,
		CRLDistributionPoints:       ca.crt.CRLDistributionPoints,
		OCSPServer:                  ca.crt.OCSPServer,
		EmailAddresses:              ca.crt.EmailAddresses,
		DNSNames:                    ca.crt.DNSNames,
		IPAddresses:                 ca.crt.IPAddresses,
		PermittedDNSDomainsCritical: ca.crt.PermittedDNSDomainsCritical,
		PermittedDNSDomains:         ca.crt.PermittedDNSDomains,
		PermittedEmailAddresses:     ca.crt.PermittedEmailAddresses,
		PermittedURIDomains:         ca.crt.PermittedURIDomains,
		PublicKeyAlgorithm:          x509.RSA,
	}

	for _, o := range opts {
		o(template)
	}

	certDER, err := x509.CreateCertificate(rand.New(rand.NewSource(time.Now().UnixNano())), template, &ca.crt, &privKey.PublicKey, &ca.privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, privKey, nil
}

func CreateX509CA(opts ...option) (*X509, error) {
	cert, privkey, err := CreateX509CACertificate(opts...)
	if err != nil {
		return nil, err
	}

	return NewX509(*cert, *privkey, *NewRand()), nil
}

func CreateX509(ca X509, opts ...option) (*X509, error) {
	cert, privkey, err := CreateX509Certificate(ca, opts...)
	if err != nil {
		return nil, err
	}

	return NewX509(*cert, *privkey, *NewRand()), nil
}

func NewX509(crt x509.Certificate, privKey rsa.PrivateKey, r rand.Rand) *X509 {
	return &X509{
		crt:     crt,
		privKey: privKey,
		r:       r,
	}
}

func ParseX509File(filename string) (*X509, error) {
	pemB, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseX509Bytes(pemB)
}

func ParseX509Bytes(pemB []byte) (*X509, error) {

	privateKeyblock, rest := pem.Decode(pemB)

	if privateKeyblock == nil || rest == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key or certificate")
	}

	certBlock, _ := pem.Decode(rest)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the certificate")
	}

	priv, err := x509.ParsePKCS1PrivateKey(privateKeyblock.Bytes)
	if err != nil {
		return nil, err
	}

	crt, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &X509{
		crt:     *crt,
		privKey: *priv,
		r:       *NewRand(),
	}, nil
}

type X509 struct {
	crt     x509.Certificate
	privKey rsa.PrivateKey
	r       rand.Rand
}

func (x *X509) Encrypt(data []byte) ([]byte, error) {
	// Encrypt data
	switch x.crt.PublicKey.(type) {
	case *rsa.PublicKey:
		return x.encryptRSA(data)
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}
}

func (x *X509) encryptRSA(data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(&x.r, x.crt.PublicKey.(*rsa.PublicKey), data)
}

func (x *X509) Decrypt(data []byte) ([]byte, error) {
	switch x.crt.PublicKey.(type) {
	case *rsa.PublicKey:
		return x.decryptRSA(data)
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
}

func (x *X509) decryptRSA(data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(&x.r, &x.privKey, data)
}

func (x *X509) GetCertificate() x509.Certificate {
	return x.crt
}

func (x *X509) PublicKey() (rsa.PublicKey, []byte) {
	pubKey := x.crt.PublicKey.(*rsa.PublicKey)
	b := x509.MarshalPKCS1PublicKey(pubKey)
	return *pubKey, b
}

func (x *X509) PrivateKey() (rsa.PrivateKey, []byte) {
	return x.privKey, x509.MarshalPKCS1PrivateKey(&x.privKey)
}

// gets the pem encoded private key and certificate in that order
func (x *X509) Pem() ([]byte, []byte) {
	keyB := x509.MarshalPKCS1PrivateKey(&x.privKey)
	keyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyB,
	}

	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: x.crt.Raw,
	}

	keyBytes := pem.EncodeToMemory(&keyBlock)
	certBytes := pem.EncodeToMemory(&certBlock)

	return keyBytes, certBytes
}
