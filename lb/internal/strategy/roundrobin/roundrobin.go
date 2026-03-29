package roundrobin

import (
	"sync/atomic"

	"go_loadbalancer/lb/internal/strategy"

	"go_loadbalancer/lb/internal/backend"
)

type RoundRobin struct {
	counter atomic.Uint64
}

func New() strategy.Strategy {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(backends []*backend.Backend) *backend.Backend {
	n := uint64(len(backends))
	if n == 0 {
		return nil
	}

	start := rr.counter.Add(1) - 1
	for i := uint64(0); i < n; i++ {
		idx := (start + i) % n
		b := backends[idx]
		if b.IsAlive() {
			return b
		}
	}

	return nil
}
