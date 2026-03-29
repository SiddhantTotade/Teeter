package gateway

import (
	"go_loadbalancer/lb/internal/handler"
	"go_loadbalancer/lb/internal/queue"
	"log"
	"net/http"
)

type Gateway struct {
	Routes  []*Route
	Handler *handler.LBHandler
}

func NewGateway(h *handler.LBHandler) *Gateway {
	return &Gateway{
		Routes:  make([]*Route, 0),
		Handler: h,
	}
}

func (g *Gateway) Register(r *Route) {
	g.Routes = append(g.Routes, r)
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range g.Routes {
		if route.Match(r.URL.Path) {
			log.Printf("Route Match: Path [%s] -> Prefix [%s]", r.URL.Path, route.Prefix)
			r.URL.Path = route.Rewrite(r.URL.Path)
			
			reqWrap := &queue.Request{
				W:        w,
				R:        r,
				Done:     make(chan struct{}),
				Registry: route.Registry,
				Strategy: route.Strategy,
				Route:    route.Prefix,
			}

			if !g.Handler.Queue.Enqueue(reqWrap) {
				http.Error(w, "Server busy", http.StatusServiceUnavailable)
				return
			}

			<-reqWrap.Done
			return
		}
	}

	http.Error(w, "route not found", http.StatusNotFound)
}
