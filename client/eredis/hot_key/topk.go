package localcache

import (
	"miopkg/util/topk"
)

func (hk *HotKey) NewTopK() *topk.Topk {
	topk := topk.NewHeavyKeeper(uint32(hk.config.LocalCacheCapacity), hk.config.width, hk.config.depth, hk.config.decay)
	return &topk
}
