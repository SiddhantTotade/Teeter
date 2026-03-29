package health

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go_loadbalancer/lb/internal/backend"
	"go_loadbalancer/lb/internal/metrics"
	"go_loadbalancer/lb/internal/registry"
)

type HealthChecker struct {
	registry         *registry.BackendRegistry
	interval         time.Duration
	failThreshold    int
	successThreshold int
	client           *http.Client
}

func NewHealthChecker(r *registry.BackendRegistry, interval time.Duration, failThreshold int, successThreshold int) *HealthChecker {
	return &HealthChecker{
		registry:         r,
		interval:         interval,
		failThreshold:    failThreshold,
		successThreshold: successThreshold,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	go func() {
		hc.check() // Check immediately on startup
		for {
			select {
			case <-ticker.C:
				hc.check()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (hc *HealthChecker) check() {
	backends := hc.registry.List()

	for _, b := range backends {
		go func(backend *backend.Backend) {
			alive := hc.ping(backend)

			// The provided code edit seems to be a replacement for the if/else block
			// and introduces new methods like SetAlive, IncrFailCount, IsAlive.
			// Assuming these methods exist on the backend.Backend struct and
			// the intent is to replace the existing health check logic with logging.
			// The original code used MarkAlive/MarkDead based on thresholds,
			// while the new code seems to directly set alive status and log.
			// I will try to integrate the logging and direct status setting as provided,
			// while keeping the threshold logic if possible, but the provided snippet
			if !alive {
				backend.IncrementFailCount()
				if backend.FailCount() >= int32(hc.failThreshold) {
					if backend.IsAlive() {
						log.Printf("Backend [%s] is now DOWN (failed %d times)", backend.URL.String(), backend.FailCount())
					}
					backend.MarkDead()
					metrics.BackendStatus.WithLabelValues(backend.RoutePrefix, backend.URL.String()).Set(0)
				}
			} else {
				backend.ResetFailCount()
				backend.IncrementSuccessCount()
				if backend.SuccessCount() >= int32(hc.successThreshold) {
					if !backend.IsAlive() {
						log.Printf("Backend [%s] is now ONLINE (succeeded %d times)", backend.URL.String(), backend.SuccessCount())
					}
					backend.MarkAlive()
					metrics.BackendStatus.WithLabelValues(backend.RoutePrefix, backend.URL.String()).Set(1)
				}
			}
		}(b)
	}
}

func (hc *HealthChecker) ping(b *backend.Backend) bool {
	req, _ := http.NewRequest("GET", b.URL.String(), nil)
	resp, err := hc.client.Do(req)

	fmt.Println("Pinging", b.URL.String(), "Alive?", err == nil)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode < 500
}
