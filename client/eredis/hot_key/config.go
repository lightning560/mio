package localcache

import (
	"miopkg/conf"
	"miopkg/log"
)

type config struct {
	Name string
	// Address is the Redis server address.
	Address string
	// LocalCache is the local cache.
	LocalCacheCapacity int
	// topk
	width uint32
	depth uint32
	decay float64
	// whiteList
	whiteList map[string]struct{}
	// client tracking
	On        bool
	Redircet  int64
	Prefix    []string
	Broadcast bool
	Optin     bool
	Optout    bool
	Noloop    bool
}

func DefaultConfig() *config {
	res := &config{
		Name:    "hotkey",
		Address: "127.0.0.1:6379",
		// topk
		width:              10000,
		depth:              5,
		decay:              0.925,
		LocalCacheCapacity: 500,
		// whiteList
		whiteList: map[string]struct{}{},
		// client tracking
		On:        true,
		Redircet:  0,
		Prefix:    []string{},
		Broadcast: true,
		Optin:     false,
		Optout:    false,
		Noloop:    false,
	}
	return res
}

func LoadConfig() *config {
	c := DefaultConfig()
	if err := conf.UnmarshalKey(c.Name, c); err != nil {
		log.Warnf("parse config error", log.FieldErr(err), log.FieldKey(c.Name))
	}
	return c
}
