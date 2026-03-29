package strategy

import "go_loadbalancer/lb/internal/backend"

type Strategy interface {
	Next([]*backend.Backend) *backend.Backend
}
