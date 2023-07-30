package localcache

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
)

func (hk *HotKey) NewLruCache() (*lru.Cache[string, interface{}], error) {
	lruCache, err := lru.New[string, interface{}](hk.config.LocalCacheCapacity)
	if err != nil {
		errors.Wrap(err, "init lru")
		return nil, err
	}
	return lruCache, nil
}

func (hk *HotKey) Add() (key string, value interface{}) {
	if ok := hk.topk.Add(key, 1); ok {
		hk.lruCache.Add(key, value)
		return
	}
	if exists := hk.inWhiteList(key); exists {
		hk.lruCache.Add(key, value)
		return
	}
	return
}

func (hk *HotKey) ContainsOrAdd(key string, value interface{}) (ok bool, evicted bool) {
	if ok := hk.topk.Add(key, 1); ok {
		return hk.lruCache.ContainsOrAdd(key, value)
	}
	if exists := hk.inWhiteList(key); exists {
		return hk.lruCache.ContainsOrAdd(key, value)
	}
	return false, false
}
func (hk *HotKey) Len() int {
	return hk.lruCache.Len()
}

func (hk *HotKey) Remove(key string) {
	hk.lruCache.Remove(key)
}

// Purge 清除 is used to completely clear the cache
func (hk *HotKey) Purge() {
	hk.lruCache.Purge()
}
func (hk *HotKey) Contains(key string) bool {
	return hk.lruCache.Contains(key)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (hk *HotKey) Keys() []string {
	return hk.lruCache.Keys()
}

func (hk *HotKey) Get(key string) (interface{}, bool) {
	return hk.lruCache.Get(key)
}
func (hk *HotKey) Peek(key string) (interface{}, bool) {
	return hk.lruCache.Peek(key)
}
