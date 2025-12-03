package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ooqls/getset/crypto/keys"
)

type TokenIssuer[C any] interface {
	IssueToken(subj string, customClaim C) (string, *jwt.Token, error)
	Decrypt(token string) (*jwt.Token, C, error)
	GetIssuer() string
}

func newParserValidator(cfg *TokenConfiguration) (*jwt.Parser, *jwt.Validator) {
	var opts []jwt.ParserOption
	for _, aud := range cfg.Audience {
		opts = append(opts, jwt.WithAudience(aud))
	}
	opts = append(opts, jwt.WithIssuer(cfg.Issuer))
	opts = append(opts, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))
	p := jwt.NewParser(opts...)
	v := jwt.NewValidator(opts...)
	return p, v
}

func NewJwtTokenIssuer[C any](cfg *TokenConfiguration,
	key keys.JwtSigningKey) TokenIssuer[C] {
	p, v := newParserValidator(cfg)
	return &jwtTokenIssuer[C]{
		key:       key,
		cfg:       cfg,
		parser:    p,
		validator: v,
	}
}

func NewDefaultJwtTokenIssuer[C any]() TokenIssuer[C] {
	tokenCfg := TokenConfiguration{
		Audience:                []string{"test"},
		Issuer:                  "default",
		IdGenType:               "uuid",
		ValidityDurationSeconds: 60 * 10,
	}

	p, v := newParserValidator(&tokenCfg)
	return &jwtTokenIssuer[C]{
		key:       keys.JWT(),
		cfg:       &tokenCfg,
		parser:    p,
		validator: v,
	}

}

type jwtTokenIssuer[C any] struct {
	key       keys.JwtSigningKey
	cfg       *TokenConfiguration
	parser    *jwt.Parser
	validator *jwt.Validator
}

func (f *jwtTokenIssuer[C]) IssueToken(subject string, customClaim C) (string, *jwt.Token, error) {
	// Check if all audience values in the token are also in our config
	if subject == "" {
		return "", nil, ErrInvalidSubject
	}

	id := uuid.New().String()
	regClaims := jwt.RegisteredClaims{
		Issuer:    f.cfg.Issuer,
		Subject:   subject,
		Audience:  f.cfg.Audience,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(f.cfg.ValidityDurationSeconds))),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        id,
	}

	claim := ClaimsWrapper[C]{
		RegisteredClaims: regClaims,
		CustomClaims:     customClaim,
	}

	return f.key.Sign(claim)
}

func (f *jwtTokenIssuer[C]) Decrypt(token string) (*jwt.Token, C, error) {

	claimWrapper := ClaimsWrapper[C]{}
	jwtToken, err := f.parser.ParseWithClaims(token, &claimWrapper, func(t *jwt.Token) (interface{}, error) {
		return f.key.PublicKey(), nil
	})

	return jwtToken, claimWrapper.CustomClaims, err
}

func (f *jwtTokenIssuer[C]) GetIssuer() string {
	return f.cfg.Issuer
}
