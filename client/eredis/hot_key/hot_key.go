package localcache

import (
	"miopkg/log"
	"miopkg/util/topk"

	lru "github.com/hashicorp/golang-lru/v2"
)

type InvalidateCallback func(keys []string)

type HotKey struct {
	lruCache *lru.Cache[string, interface{}]

	topk   topk.Topk
	config *config
}

func (hk HotKey) BuildLocalCache() *HotKey {
	config := LoadConfig()
	hk.config = config
	lruCache, err := hk.NewLruCache()
	if err != nil {
		log.Errorf("BuildLocalCache", err)
	}
	topk := hk.NewTopK()
	return &HotKey{lruCache: lruCache, topk: *topk}
}

// inWhiteList return if item is in the topk.
func (hk HotKey) inWhiteList(key string) bool {
	if _, ok := hk.config.whiteList[key]; ok {
		return true
	}
	return false
}
