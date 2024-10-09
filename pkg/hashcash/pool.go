package hashcash

import (
	"sync"
)

type HashCashPool struct {
	items map[string]*HashCash
	mu    sync.RWMutex
}

func NewHashCashPool() *HashCashPool {
	return &HashCashPool{
		items: make(map[string]*HashCash),
	}
}

func (cp *HashCashPool) AddHashCash(hc *HashCash) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.items[hc.Key()] = hc
}

func (cp *HashCashPool) RemoveHashCash(hc *HashCash) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.items, hc.Key())
}

func (cp *HashCashPool) GetHashCash(key string) (*HashCash, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	hc, exists := cp.items[key]
	return hc, exists
}
