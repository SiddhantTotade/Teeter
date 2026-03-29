package circuitbreaker

import (
	"sync/atomic"
	"time"
)

type State int32

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	FailureThreshold int32
	ResetTimeout     time.Duration
	state            atomic.Int32
	lastFailure      atomic.Int64
	failures         atomic.Int32
}

func NewCircuitBreaker(threshold int32, resetTimeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		FailureThreshold: threshold,
		ResetTimeout:     resetTimeout,
	}

	cb.state.Store(int32(Closed))
	return cb
}

func (cb *CircuitBreaker) setState(s State) {
	cb.state.Store(int32(s))
}

func (cb *CircuitBreaker) BeforeRequest() bool {
	switch cb.State() {
	case Open:
		if time.Now().UnixNano()-cb.lastFailure.Load() > cb.ResetTimeout.Nanoseconds() {
			cb.setState(HalfOpen)
			return true
		}
		return false

	case HalfOpen:
		return true
	case Closed:
		return true
	}

	return true
}

func (cb *CircuitBreaker) AfterRequestSuccess() {
	if cb.State() == HalfOpen {
		cb.setState(Closed)
	}
	cb.failures.Store(0)
}

func (cb *CircuitBreaker) AfterRequestFailure() {
	cb.failures.Add(1)
	cb.lastFailure.Store(time.Now().UnixNano())

	if cb.failures.Load() >= cb.FailureThreshold {
		cb.setState(Open)
	}
}

func (cb *CircuitBreaker) State() State {
	return State(cb.state.Load())
}
