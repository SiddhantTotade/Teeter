package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go_loadbalancer/lb/internal/backend"
	"go_loadbalancer/lb/internal/gateway"
)

type AdminServer struct {
	gateway *gateway.Gateway
}

func NewAdminServer(gw *gateway.Gateway) *AdminServer {
	return &AdminServer{
		gateway: gw,
	}
}

func (s *AdminServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/add-backend", s.handleAddBackend)
	mux.Handle("/metrics", promhttp.Handler())
	mux.ServeHTTP(w, r)
}

func (s *AdminServer) handleAddBackend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prefix string `json:"prefix"`
		URL    string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, route := range s.gateway.Routes {
		if route.Prefix == req.Prefix {
			b, err := backend.CreateNewBackend(req.Prefix, req.URL, 3*time.Second)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			route.Registry.Add(b)
			w.WriteHeader(http.StatusCreated)
			return
		}
	}

	http.Error(w, "Route not found", http.StatusNotFound)
}

func (s *AdminServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]interface{})
	
	routesStatus := make([]map[string]interface{}, 0)
	for _, route := range s.gateway.Routes {
		rs := make(map[string]interface{})
		rs["prefix"] = route.Prefix
		
		backends := make([]map[string]interface{}, 0)
		for _, b := range route.Registry.List() {
			bs := make(map[string]interface{})
			bs["url"] = b.URL.String()
			bs["alive"] = b.IsAlive()
			bs["fail_count"] = b.FailCount()
			backends = append(backends, bs)
		}
		rs["backends"] = backends
		routesStatus = append(routesStatus, rs)
	}
	
	status["routes"] = routesStatus
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
