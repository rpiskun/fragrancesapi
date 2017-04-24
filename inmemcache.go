package main

import (
	"sync"
	"time"
)

type Item struct {
	Object     interface{}
	Expiration int64
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
}

func (c *cache) Set(k string, x interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).Unix()
	}
	c.mu.Lock()
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	c.mu.Unlock()
}

func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().Unix() > item.Expiration {
			return nil, false
		}
	}
	return item.Object, true
}

func (c *cache) Delete(k string) {
	c.mu.Lock()
	delete(c.items, k)
	c.mu.Unlock()
}

func (c *cache) DeleteExpired() {
	now := time.Now().Unix()
	c.mu.Lock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

func (c *cache) Flush() {
	c.mu.Lock()
	c.items = map[string]Item{}
	c.mu.Unlock()
}

func NewCache(de time.Duration) *cache {
	items := make(map[string]Item)

	if de == DefaultExpiration {
		de = NoExpiration
	}
	c := &cache{
		defaultExpiration: de,
		items:             items,
	}

	return c
}
