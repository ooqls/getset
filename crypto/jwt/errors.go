package jwt

import "errors"

var (
	ErrInvalidAudience = errors.New("invalid audience")
	ErrInvalidSubject  = errors.New("invalid subject")
	ErrInvalidIssuer   = errors.New("invalid issuer")
) 