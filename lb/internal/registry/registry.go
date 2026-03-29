package registry

import (
	"errors"
	"sync"

	"go_loadbalancer/lb/internal/backend"
)

var (
	ErrNotFound = errors.New("backend not found")
)

type BackendRegistry struct {
	backends []*backend.Backend
	mu       sync.RWMutex
}

func NewRegistry() *BackendRegistry {
	return &BackendRegistry{
		backends: make([]*backend.Backend, 0),
	}
}

func (r *BackendRegistry) Add(b *backend.Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.backends = append(r.backends, b)
}

func (r *BackendRegistry) Remove(rawURL string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	for i, b := range r.backends {
		if b.URL.String() == rawURL {
			r.backends = append(r.backends[:i], r.backends[i+1:]...)
			return nil
		}
	}

	return ErrNotFound
}

func (r *BackendRegistry) List() []*backend.Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*backend.Backend, len(r.backends))
	copy(out, r.backends)

	return out
}

func (r *BackendRegistry) AliveBackends() []*backend.Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	alive := make([]*backend.Backend, 0)

	for _, b := range r.backends {
		if b.IsAlive() {
			alive = append(alive, b)
		}
	}

	return alive
}

func (r *BackendRegistry) MarkAlive(rawURl string, alive bool) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.backends {
		if b.URL.String() == rawURl {
			if alive {
				b.MarkAlive()
			} else {
				b.MarkDead()
			}
			return nil
		}
	}

	return ErrNotFound
}
