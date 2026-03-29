package weightedroundrobin

import (
	"sync"

	"go_loadbalancer/lb/internal/backend"
	"go_loadbalancer/lb/internal/strategy"
)

type WeightedBackend struct {
	B       *backend.Backend
	Weight  int
	Current int
}

type WeightedRoundRobin struct {
	backends []*WeightedBackend
	mu       sync.Mutex
}

func NewWeightedRoundRobin(weights map[*backend.Backend]int) strategy.Strategy {
	wrr := &WeightedRoundRobin{
		backends: make([]*WeightedBackend, 0),
	}

	for b, w := range weights {
		wrr.backends = append(wrr.backends, &WeightedBackend{
			B:      b,
			Weight: w,
		})
	}

	return wrr
}

func (wrr *WeightedRoundRobin) Next(backends []*backend.Backend) *backend.Backend {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	var candidates []*WeightedBackend

	for _, b := range backends {
		for _, wb := range wrr.backends {
			if wb.B == b && b.IsAlive() {
				wb.Current += wb.Weight
				candidates = append(candidates, wb)
				break
			}
		}
	}

	var best *WeightedBackend
	for _, wb := range candidates {
		if best == nil || wb.Current > best.Current {
			best = wb
		}
	}

	if best == nil {
		return nil
	}

	best.Current -= totalWeight(candidates)
	return best.B
}

func totalWeight(b []*WeightedBackend) int {
	sum := 0

	for _, wb := range b {
		sum += wb.Weight
	}

	return sum
}
