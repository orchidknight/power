package cache

import (
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	data map[int]string
}

func NewCache(data []string) *Cache {
	c := &Cache{
		data: make(map[int]string),
	}

	for i, s := range data {
		c.Set(i, s)
	}

	return c
}

func (c *Cache) Set(key int, value string) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
}

func (c *Cache) Get(key int) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	value, ok := c.data[key]
	return value, ok
}

func (c *Cache) GetRandom() (string, bool) {
	c.RLock()
	defer c.RUnlock()

	if len(c.data) == 0 {
		return "", false
	}

	keys := make([]int, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	randomKey := keys[rand.Intn(len(keys))]

	return c.data[randomKey], true
}
