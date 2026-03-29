package gateway

import (
	"errors"
)

var ErrRouteNotFound = errors.New("route not found")

type Router struct {
	routes []*Route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

func (r *Router) Register(route *Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) Resolve(path string) (*Route, error) {
	for _, route := range r.routes {
		if route.Match(path) {
			return route, nil
		}
	}

	return nil, ErrRouteNotFound
}
