package retry

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"go_loadbalancer/lb/internal/circuitbreaker"
	"go_loadbalancer/lb/internal/registry"
	"go_loadbalancer/lb/internal/strategy"
)

type ResponseRecorder struct {
	HeaderMap http.Header
	Body      *bytes.Buffer
	Status    int
	written   bool
}

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      &bytes.Buffer{},
		Status:    http.StatusOK,
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.HeaderMap
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.written = true
	return r.Body.Write(data)
}

func (r *ResponseRecorder) WriteHeader(status int) {
	r.Status = status
	r.written = true
}

type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	RetryOn5xx     bool
}

func DefaultPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		RetryOn5xx:     true,
	}
}

var ErrNoBackendAvailable = errors.New("no new available to serve request")

func DoWithRetries(w http.ResponseWriter, req *http.Request, reg *registry.BackendRegistry, strat strategy.Strategy, cb *circuitbreaker.CircuitBreaker, policy RetryPolicy) error {
	var bodyBuf []byte
	if req.Body != nil {
		var err error
		bodyBuf, err = io.ReadAll(req.Body)

		if err != nil {
			return fmt.Errorf("failed to read request body for retry: %w", err)
		}

		req.Body = io.NopCloser(bytes.NewReader(bodyBuf))
	}

	backends := reg.List()

	if len(backends) == 0 {
		http.Error(w, "no backends configured", http.StatusServiceUnavailable)
		return ErrNoBackendAvailable
	}

	var lastErr error
	attempted := make(map[string]bool)

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		if bodyBuf != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBuf))
		}

		alive := reg.AliveBackends()
		if len(alive) == 0 {
			http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
			return ErrNoBackendAvailable
		}

		target := strat.Next(alive)
		if target == nil {
			for _, b := range alive {
				if !attempted[b.URL.String()] {
					target = b
					break
				}
			}
		}

		if target == nil {
			target = alive[0]
		}

		if attempted[target.URL.String()] && attempt > len(alive) {
			continue
		} else {
			attempted[target.URL.String()] = true
		}

		if cb != nil {
			if !cb.BeforeRequest() {
				lastErr = fmt.Errorf("circuit open for backend %s", target.URL.String())
				target.MarkDead()
				continue
			}
		}

		rec := NewResponseRecorder()

		target.Proxy.ServeHTTP(rec, req)

		if rec.Status < 500 || !policy.RetryOn5xx {
			commitRecordedResponse(w, rec)

			target.ResetFailCount()
			if cb != nil {
				cb.AfterRequestSuccess()
			}

			return nil
		}

		lastErr = fmt.Errorf("backend %s returned status %d", target.URL.String(), rec.Status)

		target.IncrementFailCount()

		if cb != nil {
			cb.AfterRequestFailure()
		}

		if attempt == policy.MaxAttempts {
			commitRecordedResponse(w, rec)
			return lastErr
		}

		backoff := backoffDuration(policy.InitialBackoff, policy.MaxBackoff, attempt)
		time.Sleep(backoff)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("all backends had open circuits after %d attempts", policy.MaxAttempts)
	}

	return lastErr
}

func commitRecordedResponse(w http.ResponseWriter, rec *ResponseRecorder) {
	for k, vv := range rec.HeaderMap {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(rec.Status)
	_, _ = io.Copy(w, rec.Body)
}

func backoffDuration(initial, max time.Duration, attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}

	exp := math.Pow(2, float64(attempt-2))
	d := time.Duration(float64(initial) * exp)

	if d > max {
		return max
	}

	return d
}
