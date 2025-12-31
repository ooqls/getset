package app

var rsaPrivKeyPathFlag string
var rsaPubKeyPathFlag string

const (
	rsa_privateKeyPathOpt string = "opt-private-key-path"
	rsa_publicKeyPathOpt  string = "opt-public-key-path"
)

type rsaOpt struct {
	featureOpt
}

func WithPrivateKeyPath(p string) rsaOpt {
	return rsaOpt{
		featureOpt: featureOpt{
			key:   rsa_privateKeyPathOpt,
			value: p,
		},
	}
}

func WithPublicKeyPath(p string) rsaOpt {
	return rsaOpt{
		featureOpt: featureOpt{
			key:   rsa_publicKeyPathOpt,
			value: p,
		},
	}
}
func RSA(opts ...rsaOpt) RSAFeature {
	f := RSAFeature{
		Enabled:        true,
		PrivateKeyPath: rsaPrivKeyPathFlag,
		PublicKeyPath:  rsaPubKeyPathFlag,
	}
	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}

type RSAFeature struct {
	Enabled        bool
	PrivateKeyPath string
	PublicKeyPath  string
}

func (f *RSAFeature) apply(opt rsaOpt) {
	switch opt.key {
	case rsa_privateKeyPathOpt:
		f.PrivateKeyPath = opt.value.(string)
	case rsa_publicKeyPathOpt:
		f.PublicKeyPath = opt.value.(string)
	}
}
