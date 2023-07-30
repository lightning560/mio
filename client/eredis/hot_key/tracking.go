package localcache

import (
	"fmt"
	"log"

	"github.com/stfnmllr/go-resp3/client"
)

// ClientTracking is a helper function that enables Redis client-side caching
func (hk *HotKey) NewClientTracking() {
	// Create connetion providing key invalidation callback.
	dialer := new(client.Dialer)
	// 失效通知回调
	dialer.InvalidateCallback = func(keys []string) {
		for _, key := range keys {
			hk.Remove(key)
			fmt.Printf("clear localCache %s\n", key)
		}
	}
	// address "127.0.0.1:6379"
	conn, err := dialer.Dial(hk.config.Address)
	if err != nil {
		log.Fatal(err)
	}

	broadcast := true
	if err := conn.ClientTracking(true, nil, hk.config.Prefix, broadcast, false, false, false).Err(); err != nil {
		log.Fatal(err)
	}
}
