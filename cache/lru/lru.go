package lru

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

// Cache is a concurrent safe LRU cache base on "github.com/golang/groupcache/lru".
type Cache struct {
	mtx           sync.Mutex
	c             *lru.Cache
	memoryUsed    int64
	maxCachedSize int64
	onEvictedImpl func(key, value interface{})
}

type cachedNode struct {
	value interface{}
	size  int64
}

// New creates a new Cache. If maxEntries is zero, the cache has no limit and it's assumed that eviction is done by the caller.
// onEvicted optionally specificies a callback function to be executed when an entry is purged from the cache.
func NewCache(maxEntries int, maxCachedSize int64, onEvicted func(key, value interface{})) *Cache {
	c := &Cache{
		c: &lru.Cache{
			MaxEntries: maxEntries,
		},
		maxCachedSize: maxCachedSize,
	}
	if onEvicted != nil {
		c.onEvictedImpl = onEvicted
		c.c.OnEvicted = c.onEvicted
	}

	return c
}

// Add adds a value to the cache.
func (c *Cache) Add(key, value interface{}, valueSize int64) {
	c.mtx.Lock()
	c.c.Add(key, &cachedNode{value, valueSize})
	c.memoryUsed += valueSize
	for c.memoryUsed > c.maxCachedSize {
		c.c.RemoveOldest()
	}
	c.mtx.Unlock()
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.mtx.Lock()
	value, ok = c.c.Get(key)
	if ok {
		value = value.(*cachedNode).value
	}
	c.mtx.Unlock()

	return
}

// CurCachedSize returns total memory usage of the cached objects in bytes.
func (c *Cache) CurCachedSize() (size int64) {
	c.mtx.Lock()
	size = c.memoryUsed
	c.mtx.Unlock()

	return
}

// Remove removes a value from the cache.
func (c *Cache) Remove(key interface{}) {
	c.mtx.Lock()
	c.c.Remove(key)
	c.mtx.Unlock()
}

// RemoveCachedValues removes values specified in `keys` from the cache.
func (c *Cache) RemoveCachedValues(keys []interface{}) {
	c.mtx.Lock()
	for _, key := range keys {
		c.c.Remove(key)
	}
	c.mtx.Unlock()
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	c.mtx.Lock()
	c.c.Clear()
	c.mtx.Unlock()
}

func (c *Cache) onEvicted(key lru.Key, value interface{}) {
	cachedNode := value.(*cachedNode)
	c.onEvictedImpl(key, cachedNode.value)
	c.memoryUsed -= cachedNode.size
}
