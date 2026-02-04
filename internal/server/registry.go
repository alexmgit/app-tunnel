package server

import (
	"errors"
	"sync"
	"time"
)

var ErrUnknownSubdomain = errors.New("unknown subdomain")

const tunnelQueueSize = 128

type tunnelEntry struct {
	available chan *TunnelConn
	lastSeen  time.Time
}

type TunnelRegistry struct {
	store   *SubdomainStore
	timeout time.Duration
	mu      sync.Mutex
	entries map[string]*tunnelEntry
}

func NewTunnelRegistry(store *SubdomainStore, timeout time.Duration) *TunnelRegistry {
	registry := &TunnelRegistry{
		store:   store,
		timeout: timeout,
		entries: make(map[string]*tunnelEntry),
	}
	return registry
}

func (r *TunnelRegistry) RegisterSubdomain(requested string) (string, error) {
	subdomain, err := r.store.Register(requested)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[subdomain]; !exists {
		r.entries[subdomain] = &tunnelEntry{
			available: make(chan *TunnelConn, tunnelQueueSize),
			lastSeen:  time.Now(),
		}
	}

	return subdomain, nil
}

func (r *TunnelRegistry) RegisterTunnel(subdomain string, conn *TunnelConn) error {
	r.mu.Lock()
	entry, ok := r.entries[subdomain]
	if !ok {
		r.mu.Unlock()
		return ErrUnknownSubdomain
	}
	entry.lastSeen = time.Now()
	r.mu.Unlock()

	select {
	case entry.available <- conn:
		return nil
	default:
		return errors.New("tunnel queue full")
	}
}

func (r *TunnelRegistry) Acquire(subdomain string, timeout time.Duration) (*TunnelConn, error) {
	r.mu.Lock()
	entry, ok := r.entries[subdomain]
	if !ok {
		r.mu.Unlock()
		return nil, ErrUnknownSubdomain
	}
	entry.lastSeen = time.Now()
	r.mu.Unlock()

	select {
	case conn := <-entry.available:
		return conn, nil
	case <-time.After(timeout):
		return nil, errors.New("no tunnel available")
	}
}

func (r *TunnelRegistry) ReleaseInactive() {
	now := time.Now()
	var stale []string

	r.mu.Lock()
	for subdomain, entry := range r.entries {
		if now.Sub(entry.lastSeen) > r.timeout {
			stale = append(stale, subdomain)
		}
	}
	for _, subdomain := range stale {
		delete(r.entries, subdomain)
		r.store.Release(subdomain)
	}
	r.mu.Unlock()
}

func (r *TunnelRegistry) StartReaper(stop <-chan struct{}) {
	ticker := time.NewTicker(r.timeout / 2)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.ReleaseInactive()
			case <-stop:
				return
			}
		}
	}()
}
