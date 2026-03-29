package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"go_loadbalancer/lb/internal/backend"
	"go_loadbalancer/lb/internal/queue"
	"go_loadbalancer/lb/internal/ratelimit"
	"go_loadbalancer/lb/internal/registry"
	"go_loadbalancer/lb/internal/retry"
	"go_loadbalancer/lb/internal/strategy"
)

type LBHandler struct {
	Registry      *registry.BackendRegistry
	Strategy      strategy.Strategy
	MaxRetries    int
	GlobalLimiter *ratelimit.TokenBucket
	Queue         *queue.RequestQueue
}

func NewHandler(r *registry.BackendRegistry, s strategy.Strategy, maxRetries int, q *queue.RequestQueue) *LBHandler {
	h := &LBHandler{
		Registry:   r,
		Strategy:   s,
		MaxRetries: maxRetries,
		Queue:      q,
	}

	return h
}

func (h *LBHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	reqWrap := &queue.Request{
		W:        w,
		R:        req,
		Done:     make(chan struct{}),
		Registry: h.Registry,
		Strategy: h.Strategy,
	}

	if h.GlobalLimiter != nil && !h.GlobalLimiter.Allow() {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	if !h.Queue.Enqueue(reqWrap) {
		http.Error(w, "Server busy. Too many requests.", http.StatusServiceUnavailable)
		return
	}

	<-reqWrap.Done

}

func (h *LBHandler) processRequest(w http.ResponseWriter, req *http.Request, reg *registry.BackendRegistry, strat strategy.Strategy) {
	var bodyBuf []byte
	if req.Body != nil {
		if b, err := io.ReadAll(req.Body); err == nil {
			bodyBuf = b
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBuf))
	}

	var lastErr error

	for attempt := 0; attempt < h.MaxRetries; attempt++ {

		alive := reg.AliveBackends()
		if len(alive) == 0 {
			http.Error(w, "no backend available", http.StatusServiceUnavailable)
			return
		}

		backend := strat.Next(alive)
		if backend == nil {
		log.Printf("No healthy backends available for path: %s", req.URL.Path)
		http.Error(w, fmt.Sprintf("No healthy backend service found for path %s. Please check if your services are running.", req.URL.Path), http.StatusServiceUnavailable)
		return
	}

		if backend.CB != nil {
			if ok := backend.CB.BeforeRequest(); !ok {
				log.Printf("circuit OPEN for %s → skipping", backend.URL)
				continue
			}
		}

		if bodyBuf != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBuf))
		}

		if isWebSocket(req) {
			h.proxyWebSocket(w, req, backend)
			return
		}

		log.Printf("Forwarding %s %s to %s", req.Method, req.URL.Path, backend.URL)
		rec := retry.NewResponseRecorder()
		backend.Proxy.ServeHTTP(rec, req)
		log.Printf("Backend returned %d with %d headers", rec.Status, len(rec.HeaderMap))

		if rec.Status < 500 {
			backend.RecordSuccess()

			for k, vv := range rec.HeaderMap {
				for _, v := range vv {
					if strings.Contains(strings.ToLower(k), "cookie") {
						log.Printf("Passing header: %s: %s", k, v)
					}
					w.Header().Add(k, v)
				}
			}

			w.WriteHeader(rec.Status)

			if rec.Body != nil {
				_, _ = io.Copy(w, rec.Body)
			}

			log.Printf(
				"backend SUCCESS: %s status=%d attempt=%d",
				backend.URL, rec.Status, attempt+1,
			)

			return
		}

		lastErr = fmt.Errorf("backend %s returned %d", backend.URL, rec.Status)
		backend.RecordFailure()

		log.Printf(
			"backend FAILURE: %s status=%d attempt=%d",
			backend.URL, rec.Status, attempt+1,
		)

		time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("all backend failed")
	}

	http.Error(w, lastErr.Error(), http.StatusBadGateway)
}

func (h *LBHandler) ServeBackend(w http.ResponseWriter, r *http.Request, reg *registry.BackendRegistry, strat strategy.Strategy) {
	h.processRequest(w, r, reg, strat)
}

func isWebSocket(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

func (h *LBHandler) proxyWebSocket(w http.ResponseWriter, r *http.Request, b *backend.Backend) {
	targetURL := *b.URL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery
	if targetURL.Scheme == "http" {
		targetURL.Scheme = "ws"
	} else if targetURL.Scheme == "https" {
		targetURL.Scheme = "wss"
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	backConn, err := net.Dial("tcp", b.URL.Host)
	if err != nil {
		http.Error(w, "could not connect to backend", http.StatusServiceUnavailable)
		return
	}
	defer backConn.Close()

	if err := r.Write(backConn); err != nil {
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(backConn, conn)
	go cp(conn, backConn)
	<-errc
}
