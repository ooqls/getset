package crypto

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPcg32(t *testing.T) {
	tSeed := time.Unix(1000, 1000)
	tSeed2 := tSeed.Add(1 * time.Minute)
	PCG32Seed(uint64(tSeed.UnixMilli()), uint64(tSeed2.UnixMilli()))

	rng := NewPCG32(uint64(tSeed.UnixMilli()), uint64(tSeed2.UnixMilli()))
	assert.NotNil(t, rng, "should be able to create rng")

	assert.NotZerof(t, rng.Next(), "should be able to generate number")

	for i := 0; i < 5; i++ {
		log.Printf("%d", rng.Next())
	}

	buff := make([]byte, 20)
	PCG32Read(buff)
	notZero := false
	for i := range 20 {
		if buff[i] != 0 {
			notZero = true
			break
		}
	}
	assert.True(t, notZero, "should be able to read")

	buff2 := make([]byte, 20)
	rng.Read(buff2)

	notZero = false
	for i := range 20 {
		if buff2[i] != 0 {
			notZero = true
			break
		}
	}
	assert.True(t, notZero, "should be able to read")
}
