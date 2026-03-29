package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"go_loadbalancer/lb/internal/circuitbreaker"
)

type Backend struct {
	URL       *url.URL
	Proxy     *httputil.ReverseProxy
	Alive     atomic.Bool
	Transport *http.Transport
	Failures  atomic.Int32
	Successes atomic.Int32

	CB *circuitbreaker.CircuitBreaker
}

func CreateNewBackend(rawURL string, timeout time.Duration) (*Backend, error) {
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalHost := req.Host
		originalDirector(req)
		// We don't overwrite req.Host here anymore, keeping it as the backend's host
		// but providing the original info in X-Forwarded headers.
		req.Header.Set("X-Forwarded-Host", originalHost)
		req.Header.Set("X-Forwarded-Proto", "http")
	}
	proxy.Transport = transport

	b := &Backend{
		URL:       parsed,
		Proxy:     proxy,
		Transport: transport,
		CB:        circuitbreaker.NewCircuitBreaker(3, 5*time.Second),
	}

	b.Alive.Store(true)

	return b, nil
}

func (b *Backend) MarkAlive() {
	b.Alive.Store(true)
}

func (b *Backend) MarkDead() {
	b.Alive.Store(false)
}

func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}

func (b *Backend) IncrementFailCount() { b.Failures.Add(1) }
func (b *Backend) ResetFailCount()     { b.Failures.Store(0) }
func (b *Backend) FailCount() int32    { return b.Failures.Load() }

func (b *Backend) IncrementSuccessCount() { b.Successes.Add(1) }
func (b *Backend) ResetSuccessCount()     { b.Successes.Store(0) }
func (b *Backend) SuccessCount() int32    { return b.Successes.Load() }

func (b *Backend) RecordSuccess() {
	b.IncrementSuccessCount()
	b.CB.AfterRequestSuccess()
}

func (b *Backend) RecordFailure() {
	b.IncrementFailCount()
	b.CB.AfterRequestFailure()

	if b.FailCount() >= b.CB.FailureThreshold {
		b.MarkDead()
	}
}
