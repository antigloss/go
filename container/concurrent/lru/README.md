lru
======

Package lru provides a goroutine safe LRU cache implementation based on "github.com/golang/groupcache/lru".

### Usage

``` 
// Creates a ready-to-use cache
cache := lru.NewCache(MaxCachedFileNum, MaxCachedSize, func(key, object interface{}) {
	// Jobs to do on evicted
})
// Caches an object
cache.Add(Key, CachedObj, CachedObjSize)
// Gets a cached object
cachedObj, ok := cache.Get(Key)
```
