package main

import (
	"sync"
	"time"
)

type Cache struct {
	data map[string]cacheEntry
	lock *sync.Mutex
}

func (c *Cache) Add(key string, value []byte) {
	entry := cacheEntry{
		createdAt: time.Now(),
		val:       value,
	}
	c.data[key] = entry
}
func (c *Cache) Get(key string) ([]byte, bool) {
	if entry, ok := c.data[key]; ok {
		return entry.val, true
	}
	return nil, false
}
func (c *Cache) reapLoop(interval time.Duration) {
	for k, v := range c.data {
		if time.Since(v.createdAt) > interval {
			c.lock.Lock()
			delete(c.data, k)
			c.lock.Unlock()
		}
	}
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	ticker := time.NewTicker(interval)
	cache := Cache{
		data: make(map[string]cacheEntry),
		lock: &sync.Mutex{},
	}
	go func() {
		for range ticker.C {
			cache.reapLoop(interval)
		}
	}()

	return cache
}

type PokeMap struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}
