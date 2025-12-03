package crypto

import "errors"

var (
	ErrDataTooShort = errors.New("data too short")
	ErrInvalidData  = errors.New("invalid data")
)
