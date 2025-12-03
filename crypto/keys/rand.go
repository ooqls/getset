package keys

import (
	"math/rand"
	"time"
)

func NewRand(seeds ...int64) *rand.Rand {
	seed := time.Now().UnixNano()
	for s := range seeds {
		seed += int64(s)
	}

	return rand.New(rand.NewSource(seed))
}
