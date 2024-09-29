package server

import "sync"

type ClientPool struct {
	clients map[string]struct{}
	mu      sync.RWMutex
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[string]struct{}),
	}
}

func (cp *ClientPool) AddClient(ip string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.clients[ip] = struct{}{}
}

func (cp *ClientPool) RemoveClient(ip string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.clients, ip)
}

func (cp *ClientPool) HasClient(ip string) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	_, exists := cp.clients[ip]
	return exists
}
