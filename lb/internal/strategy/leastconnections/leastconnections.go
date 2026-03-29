package leastconnections

import (
	"go_loadbalancer/lb/internal/backend"
	"sync"
)

type LeastConnections struct {
	mu          sync.Mutex
	connections map[*backend.Backend]int
}

func NewLeastConnections() *LeastConnections {
	return &LeastConnections{
		connections: make(map[*backend.Backend]int),
	}
}

func (lc *LeastConnections) Next(backends []*backend.Backend) *backend.Backend {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if len(backends) == 0 {
		return nil
	}

	var chosen *backend.Backend
	minConn := int(^uint(0) >> 1)

	for _, b := range backends {
		if !b.IsAlive() {
			continue
		}

		if _, exists := lc.connections[b]; !exists {
			lc.connections[b] = 0
		}

		if lc.connections[b] < minConn {
			minConn = lc.connections[b]
			chosen = b
		}
	}

	if chosen != nil {
		lc.connections[chosen]++
	}

	return chosen

}

func (lc *LeastConnections) Done(b *backend.Backend) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.connections[b] > 0 {
		lc.connections[b]--
	}
}
