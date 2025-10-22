package internal

import (
	"sync"
	"time"
)

var mu sync.Mutex

type Cache struct {
	entryList map[string]cacheEntry
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func (c Cache) Add(key string, val []byte) {
	mu.Lock()
	defer mu.Unlock()
	c.entryList[key] = cacheEntry{createdAt: time.Now(), val: val}
}

func (c Cache) Get(key string) ([]byte, bool) {
	mu.Lock()
	defer mu.Unlock()
	entry, exists := c.entryList[key]
	if !exists {
		return nil, false
	}
	return entry.val, true
}

func (c Cache) reapLoop(interval time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	for key, entry := range c.entryList {
		if time.Since(entry.createdAt) > interval {
			delete(c.entryList, key)
		}
	}
}

func NewCache(interval time.Duration) Cache {
	theCache := Cache{entryList: map[string]cacheEntry{}}
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			theCache.reapLoop(interval)
		}
	}()
	return theCache
}
