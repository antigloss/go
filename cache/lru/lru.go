/*
 *
 * lru - LRU cache package
 * Copyright (C) 2018 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

/*
Package lru provides a goroutine safe LRU cache implementation based on "github.com/golang/groupcache/lru".

Basic example:

	// Creates a cache
	cache := lru.NewCache(MaxCachedFileNum, MaxCachedSize, func(key, object interface{}) {
		// Jobs to do on evicted
	})
	// Caches an object
	cache.Add(Key, CachedObj, CachedObjSize)
	// Gets a cached object
	cachedObj, ok := cache.Get(Key)
*/
package lru

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

// Cache is a goroutine safe LRU cache base on "github.com/golang/groupcache/lru".
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

// NewCache creates a ready-to-use Cache.
//
//	maxEntries: Limit of cached objects, LRU eviction will be triggered when reached.
//	maxCachedSize: Limit of total cached objects' size in bytes, LRU eviction will be triggered when reached.
//	onEvicted: Optionally specificies a callback function to be executed when an entry is purged from the cache.
func NewCache(maxEntries int, maxCachedSize int64, onEvicted func(key, object interface{})) *Cache {
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

// Add adds an object to the cache, LRU eviction will be triggered if limit reached after adding.
//
//	key: Key of the cached object.
//	object: Object to be cached.
//	objectSize: Size in bytes of the cached object.
func (c *Cache) Add(key, object interface{}, objectSize int64) {
	c.mtx.Lock()
	c.c.Add(key, &cachedNode{object, objectSize})
	c.memoryUsed += objectSize
	for c.memoryUsed > c.maxCachedSize {
		c.c.RemoveOldest()
	}
	c.mtx.Unlock()
}

// Get looks up a key's object from the cache. It returns true and the object if found, false and nil otherwise.
func (c *Cache) Get(key interface{}) (object interface{}, ok bool) {
	c.mtx.Lock()
	object, ok = c.c.Get(key)
	if ok {
		object = object.(*cachedNode).value
	}
	c.mtx.Unlock()

	return
}

// CurCachedSize returns the total cached objects' size in bytes.
func (c *Cache) CurCachedSize() (size int64) {
	c.mtx.Lock()
	size = c.memoryUsed
	c.mtx.Unlock()

	return
}

// Remove removes a key's object from the cache.
func (c *Cache) Remove(key interface{}) {
	c.mtx.Lock()
	c.c.Remove(key)
	c.mtx.Unlock()
}

// RemoveCachedObjects removes objects specified in `keys` from the cache.
func (c *Cache) RemoveCachedObjects(keys []interface{}) {
	c.mtx.Lock()
	for _, key := range keys {
		c.c.Remove(key)
	}
	c.mtx.Unlock()
}

// Clear purges all cached objects from the cache.
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
