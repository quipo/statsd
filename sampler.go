package statsd

import (
	"math/rand"
	"sync"
	"time"
)

// SyncSampler is a sampler designed for non-contended cases.
type SyncSampler struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// PoolSampler is a sampler designed for contended cases.
type PoolSampler struct {
	pool sync.Pool
}

var DefaultSampler = NewPoolSampler()

func NewSyncSampler() *SyncSampler {
	return &SyncSampler{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func NewPoolSampler() *PoolSampler {
	return &PoolSampler{
		pool: sync.Pool{
			New: func() interface{} {
				return rand.New(rand.NewSource(time.Now().UnixNano()))
			},
		},
	}
}

func (s *SyncSampler) ShouldFire(rate float32) bool {
	s.mu.Lock()
	shouldFire := roll(s.rng, rate)
	s.mu.Unlock()
	return shouldFire
}

func (s *PoolSampler) ShouldFire(rate float32) bool {
	rng := s.pool.Get().(*rand.Rand)
	shouldFire := roll(rng, rate)
	s.pool.Put(rng)
	return shouldFire
}

func roll(rng *rand.Rand, sampleRate float32) bool {
	return rng.Float32() <= sampleRate
}
