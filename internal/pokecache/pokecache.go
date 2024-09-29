package pokecache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	cachemap map[string]CacheEntry
	mu       sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	return &Cache{
		cachemap: make(map[string]CacheEntry),
	}
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := CacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.cachemap[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, found := c.cachemap[key]
	if !found {
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.mu.Lock()
		for key, entry := range c.cachemap {
			if time.Since(entry.createdAt) > interval {
				delete(c.cachemap, key)
			}
		}
		c.mu.Unlock()
	}
}
