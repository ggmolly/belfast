package misc

import (
	"encoding/binary"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type LockedRand struct {
	mu sync.Mutex
	r  *rand.Rand
}

func NewLockedRand() *LockedRand {
	return &LockedRand{r: rand.New(rand.NewChaCha8(chaCha8Seed()))}
}

func (r *LockedRand) Uint32N(n uint32) uint32 {
	r.mu.Lock()
	value := r.r.Uint32N(n)
	r.mu.Unlock()
	return value
}

var rngSeedCounter atomic.Uint64

func chaCha8Seed() [32]byte {
	var seed [32]byte
	value := uint64(time.Now().UnixNano()) ^ rngSeedCounter.Add(1)
	for i := 0; i < len(seed); i += 8 {
		value = splitMix64(value)
		binary.LittleEndian.PutUint64(seed[i:], value)
	}
	return seed
}

func splitMix64(value uint64) uint64 {
	// SplitMix64 mixing step to expand a single counter into 64-bit seed material.
	value += 0x9e3779b97f4a7c15
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}
