package gateway

import (
	"fmt"
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
				handler.ServeErrorPage(w, r, http.StatusServiceUnavailable, "Server Busy", "The load balancer queue is currently full. Please try again in a few seconds.")
				return
			}

			<-reqWrap.Done
			return
		}
	}

	handler.ServeErrorPage(w, r, http.StatusNotFound, "Route Not Found", fmt.Sprintf("The requested path '%s' is not registered in our routing configuration.", r.URL.Path))
}
