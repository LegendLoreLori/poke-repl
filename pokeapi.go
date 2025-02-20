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

type PokeMapResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type PokeMap struct { // expand with more fields later? encounter method etc
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	caught  int
	ID      int `json:"id"`
	Height  int `json:"height"`
	Weight  int `json:"weight"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
}

// type PokedexEntry struct {
// 	ID                   int    `json:"id"`
// 	IsBaby               bool   `json:"is_baby"`
// 	IsLegendary          bool   `json:"is_legendary"`
// 	IsMythical           bool   `json:"is_mythical"`
// 	Name                 string `json:"name"`
// 	Shape struct {
// 		Name string `json:"name"`
// 		URL  string `json:"url"`
// 	} `json:"shape"`
// 	Varieties []struct {
// 		IsDefault bool `json:"is_default"`
// 		Pokemon   struct {
// 			Name string `json:"name"`
// 			URL  string `json:"url"`
// 		} `json:"pokemon"`
// 	} `json:"varieties"`
// }
