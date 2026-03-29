package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"go_loadbalancer/lb/internal/admin"
	"go_loadbalancer/lb/internal/backend"
	"go_loadbalancer/lb/internal/gateway"
	"go_loadbalancer/lb/internal/handler"
	"go_loadbalancer/lb/internal/health"
	"go_loadbalancer/lb/internal/queue"
	"go_loadbalancer/lb/internal/registry"
	"go_loadbalancer/lb/internal/strategy"
	"go_loadbalancer/lb/internal/strategy/leastconnections"
	"go_loadbalancer/lb/internal/strategy/roundrobin"
	"go_loadbalancer/lb/internal/strategy/weightedroundrobin"
	"go_loadbalancer/lb/pkg/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	q := queue.NewRequestQueue(100)
	h := handler.NewHandler(nil, nil, 3, q)
	gw := gateway.NewGateway(h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, rc := range cfg.Routes {
		reg := registry.NewRegistry()
		weights := make(map[*backend.Backend]int)

		for _, bc := range rc.Backends {
			timeout, _ := time.ParseDuration(bc.Timeout)
			if timeout == 0 {
				timeout = 3 * time.Second
			}
			b, err := backend.CreateNewBackend(bc.URL, timeout)
			if err != nil {
				log.Printf("failed to create backend %s: %v", bc.URL, err)
				continue
			}
			reg.Add(b)
			if bc.Weight > 0 {
				weights[b] = bc.Weight
			}
		}

		var strat strategy.Strategy
		switch rc.Strategy {
		case "least_connections":
			strat = leastconnections.NewLeastConnections()
		case "weighted_round_robin":
			strat = weightedroundrobin.NewWeightedRoundRobin(weights)
		default:
			strat = roundrobin.New()
		}

		gw.Register(&gateway.Route{
			Prefix:   rc.Prefix,
			Registry: reg,
			Strategy: strat,
		})

		hc := health.NewHealthChecker(reg, 2*time.Second, 3, 2)
		hc.Start(ctx)
	}

	q.StartWorkers(50, func(req *queue.Request) {
		h.ServeBackend(req.W, req.R, req.Registry, req.Strategy)
	})

	if cfg.Admin > 0 {
		adminSrv := admin.NewAdminServer(gw)
		adminAddr := fmt.Sprintf(":%d", cfg.Admin)
		go func() {
			log.Printf("Admin API running on %s", adminAddr)
			if err := http.ListenAndServe(adminAddr, adminSrv); err != nil {
				log.Printf("Admin server failed: %v", err)
			}
		}()
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Load Balancer running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, gw))
}
