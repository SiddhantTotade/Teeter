package gateway

import (
	"go_loadbalancer/lb/internal/registry"
	"go_loadbalancer/lb/internal/strategy"
	"strings"
)

type Route struct {
	Prefix       string
	StringPrefix bool
	Registry     *registry.BackendRegistry
	Strategy     strategy.Strategy
}

func (r *Route) Match(path string) bool {
	return strings.HasPrefix(path, r.Prefix)
}

func (r *Route) Rewrite(path string) string {
	if r.StringPrefix {
		return strings.TrimPrefix(path, r.Prefix)
	}

	return path
}
