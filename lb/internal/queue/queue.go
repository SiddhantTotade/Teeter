package queue

import (
	"net/http"
	"go_loadbalancer/lb/internal/registry"
	"go_loadbalancer/lb/internal/strategy"
)

type Request struct {
	W        http.ResponseWriter
	R        *http.Request
	Done     chan struct{}
	Registry *registry.BackendRegistry
	Strategy strategy.Strategy
	Route    string
}

type RequestQueue struct {
	queue chan *Request
}

func NewRequestQueue(maxQueueSize int) *RequestQueue {
	return &RequestQueue{
		queue: make(chan *Request, maxQueueSize),
	}
}

func (rq *RequestQueue) Enqueue(req *Request) bool {
	select {
	case rq.queue <- req:
		return true
	default:
		return false
	}
}

func (rq *RequestQueue) StartWorkers(workerCount int, handler func(r *Request)) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for req := range rq.queue {
				handler(req)
				if req.Done != nil {
					close(req.Done)
				}
			}
		}()
	}
}
