package ssh

import (
	"fmt"
	"sync"
)

// Pool manages a set of named SSH clients, keyed by host name.
type Pool struct {
	mu      sync.Mutex
	clients map[string]*Client
}

// NewPool creates an empty connection pool.
func NewPool() *Pool {
	return &Pool{clients: make(map[string]*Client)}
}

// Get returns an existing client for the given name, or opens a new one.
func (p *Pool) Get(name string, cfg Config) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if c, ok := p.clients[name]; ok {
		return c, nil
	}

	c, err := Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("pool connect %s: %w", name, err)
	}
	p.clients[name] = c
	return c, nil
}

// CloseAll terminates every connection held by the pool.
func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, c := range p.clients {
		_ = c.Close()
		delete(p.clients, name)
	}
}

// Remove closes and removes a single client from the pool.
func (p *Pool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if c, ok := p.clients[name]; ok {
		_ = c.Close()
		delete(p.clients, name)
	}
}

// Len returns the number of active clients currently held in the pool.
func (p *Pool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.clients)
}

// Names returns the names of all clients currently held in the pool.
func (p *Pool) Names() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	names := make([]string, 0, len(p.clients))
	for name := range p.clients {
		names = append(names, name)
	}
	return names
}
