package crypto

import (
	"encoding/binary"
	"sync"
)

var (
	state uint64 = 0
	inc   uint64 = 0
)

var m *sync.Mutex = &sync.Mutex{}

func init() {
	m.Lock()
	defer m.Unlock()

	state = 0
	inc = 0
}

func PCG32Seed(seed, seq uint64) {
	m.Lock()
	defer m.Unlock()

	state = 0
	inc = (seq << 1) | 1
	next()
	state += seed
	next()
}

func PCG32Read(buff []byte) {
	m.Lock()
	defer m.Unlock()

	for i := 0; i < len(buff); i += 4 {
		var temp [4]byte
		binary.LittleEndian.PutUint32(temp[:], next())
		copy(buff[i:], temp[:])
	}
}
func next() uint32 {
	oldstate := state
	state = oldstate*6364136223846793005 + inc
	xorshifted := uint32(((oldstate >> 18) ^ oldstate) >> 27)
	rot := uint32(oldstate >> 59)
	return (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
}

type PCG32 struct {
	state uint64
	inc   uint64
	m     *sync.Mutex
}

func NewPCG32(seed, seq uint64) *PCG32 {
	p := &PCG32{
		m: &sync.Mutex{},
		inc: (seq << 1) | 1,
		state: 0,
	}
	p.Seed(seed, seq)
	return p
}

func (p *PCG32) Seed(seed, seq uint64) {
	p.m.Lock()
	defer p.m.Unlock()

	p.state = 0
	p.inc = (seq << 1) | 1
	p.Next()
	p.state += seed
	p.Next()
}

func (p *PCG32) Read(buff []byte) int {
	for i := 0; i < len(buff); i += 4 {
		var temp [4]byte
		binary.LittleEndian.PutUint32(temp[:], p.Next())
		copy(buff[i:], temp[:])
	}

	return len(buff)
}

func (p *PCG32) Next() uint32 {
	oldstate := p.state
	p.state = oldstate*6364136223846793005 + p.inc
	xorshifted := uint32(((oldstate >> 18) ^ oldstate) >> 27)
	rot := uint32(oldstate >> 59)
	return (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
}
