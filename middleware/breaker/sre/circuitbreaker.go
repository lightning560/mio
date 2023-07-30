package sre

import (
	"errors"
	"sync"
)

// ErrNotAllowed error not allowed.
var ErrNotAllowed = errors.New("circuitbreaker: not allowed for circuit open")

// CircuitBreaker is a circuit breaker.
type CircuitBreaker interface {
	Allow() error
	MarkSuccess()
	MarkFailed()
}

// Group represents a class of CircuitBreaker and forms a namespace in which
// units of CircuitBreaker.
type Group struct {
	mu   sync.RWMutex
	brks map[string]CircuitBreaker
}

// NewGroup new a breaker group container, if conf nil use default conf.
func NewGroup() *Group {

	return &Group{
		brks: make(map[string]CircuitBreaker),
	}
}

// Get get a breaker by a specified key, if breaker not exists then make a new one.
func (g *Group) Get(key string) CircuitBreaker {
	g.mu.RLock()
	brk, ok := g.brks[key]
	g.mu.RUnlock()
	if ok {
		return brk
	}
	// NOTE here may new multi breaker for rarely case, let gc drop it.
	brk = NewBreaker()
	g.mu.Lock()
	if _, ok = g.brks[key]; !ok {
		g.brks[key] = brk
	}
	g.mu.Unlock()
	return brk
}

// Reload reload the group by specified config, this may let all inner breaker
// reset to a new one.
func (g *Group) Reload(conf *Breaker) {
	if conf == nil {
		return
	}
	g.mu.Lock()
	g.brks = make(map[string]CircuitBreaker, len(g.brks))
	g.mu.Unlock()
}

// Go runs your function while tracking the breaker state of group.
func (g *Group) Go(name string, run, fallback func() error) error {
	breaker := g.Get(name)
	if err := breaker.Allow(); err != nil {
		return fallback()
	}
	return run()
}
