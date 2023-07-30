package etcdv3

import (
	"miopkg/registry"
)

func init() {
	registry.RegisterBuilder("etcdv3", func(confKey string) registry.Registry {
		return RawConfig(confKey).MustBuild()
	})
}
