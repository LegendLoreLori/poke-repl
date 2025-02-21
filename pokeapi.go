package main

import (
	"fmt"
	"strings"
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
		VersionDetails []struct {
			Version struct {
				Name string `json:"name"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	ID             int `json:"id"`
	Height         int `json:"height"`
	Weight         int `json:"weight"`
	BaseExperience int `json:"base_experience"`
	Species        struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	PokedexEntry
}

type PokedexEntry struct {
	caught            int
	encountered       int
	IsLegendary       bool `json:"is_legendary"`
	IsMythical        bool `json:"is_mythical"`
	FlavorTextEntries []struct {
		FlavorText string `json:"flavor_text"`
		Language   struct {
			Name string `json:"name"`
		} `json:"language"`
		Version struct {
			Name string `json:"name"`
		} `json:"version"`
	} `json:"flavor_text_entries"`
	Color struct {
		Name string `json:"name"`
	} `json:"color"`
	Shape struct {
		Name string `json:"name"`
	} `json:"shape"`
}

func (p *PokedexEntry) GetDexEntry(locale string, generation string) (string, error) {
	for _, entry := range p.FlavorTextEntries {
		if entry.Language.Name == locale && entry.Version.Name == generation {
			formattedEntry := strings.Join(strings.Fields(entry.FlavorText), " ")
			return formattedEntry, nil
		}
	}
	return "", fmt.Errorf("unable to locate pokedex entry with locale: %s version: %s", locale, generation)
}

type Pokeball struct {
	catchModifier int
	name          string
}
