package app

import "github.com/ooqls/getset/crypto/jwt"

var jwtPrivKeyPathFlag string
var jwtPubKeyPathFlag string

type jwtOpt struct {
	featureOpt
}

const (
	jwt_tokenConfigurationPathOpt string = "jwt_tokenConfigurationPath"
	jwt_tokenConfigurationOpt     string = "jwt_tokenConfiguration"
	jwt_privateKeyPathOpt         string = "jwt_privateKeyPath"
	jwt_publicKeyPathOpt          string = "jwt_publicKeyPath"
)

func WithTokenConfigurationPaths(p []string) jwtOpt {
	return jwtOpt{
		featureOpt: featureOpt{
			key:   jwt_tokenConfigurationPathOpt,
			value: p,
		},
	}
}

func WithTokenConfigurations(cfg []jwt.TokenConfiguration) jwtOpt {
	return jwtOpt{
		featureOpt: featureOpt{
			key:   jwt_tokenConfigurationOpt,
			value: cfg,
		},
	}
}

func WithJWTPrivateKeyPath(p string) jwtOpt {
	return jwtOpt{
		featureOpt: featureOpt{
			key:   jwt_privateKeyPathOpt,
			value: p,
		},
	}
}

func WithJWTPublicKeyPath(p string) jwtOpt {
	return jwtOpt{
		featureOpt: featureOpt{
			key:   jwt_publicKeyPathOpt,
			value: p,
		},
	}
}

func JWT(opts ...jwtOpt) JWTFeature {
	f := JWTFeature{
		Enabled:        true,
		PrivateKeyPath: jwtPrivKeyPathFlag,
		PubKeyPath:     jwtPubKeyPathFlag,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}

type JWTFeature struct {
	Enabled                 bool
	PrivateKeyPath          string
	PubKeyPath              string
	tokenConfigurationPaths []string
	tokenConfiguration      []jwt.TokenConfiguration
}

func (f *JWTFeature) apply(opt jwtOpt) {
	switch opt.key {
	case jwt_privateKeyPathOpt:
		f.PrivateKeyPath = opt.value.(string)
	case jwt_publicKeyPathOpt:
		f.PubKeyPath = opt.value.(string)
	case jwt_tokenConfigurationPathOpt:
		f.tokenConfigurationPaths = opt.value.([]string)
	case jwt_tokenConfigurationOpt:
		f.tokenConfiguration = opt.value.([]jwt.TokenConfiguration)
	}
}
