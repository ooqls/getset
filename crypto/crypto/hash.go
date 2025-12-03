package crypto

import (
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/argon2"
)

var h hash.Hash = sha512.New()

func GetDefaultHash() hash.Hash {
	return h
}

func SetDefaultHash(newh hash.Hash) {
	h = newh
}

type HashingAlgorithm interface {
	Hash(data []byte) []byte
}

func NewArgon2HashingAlgorithm(memory uint32, time uint32, threads uint8, keyLen uint32, salt []byte) HashingAlgorithm {
	return &Argon2HashingAlgorithm{
		Memory:  memory,
		Time:    time,
		Threads: threads,
		KeyLen:  keyLen,
		Salt:    salt,
	}
}

type Argon2HashingAlgorithm struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	KeyLen  uint32
	Salt    []byte
}

func (a *Argon2HashingAlgorithm) Hash(data []byte) []byte {
	return argon2.IDKey(data, []byte(a.Salt), a.Memory, a.Time, a.Threads, a.KeyLen)
}
