package eredbloom

import (
	"miopkg/conf"
	"miopkg/log"
)

type config struct {
	Name         string
	Address      string
	RedbloomName string
	AuthPass     *string
}

func DefaultConfig() *config {
	return &config{
		Name:         "redbloom",
		Address:      "127.0.0.1:6379",
		RedbloomName: "nohelp",
		AuthPass:     nil,
	}
}
func LoadConfig(key string) *config {
	c := DefaultConfig()
	if err := conf.UnmarshalKey(key, c); err != nil {
		log.Error("parse config error", log.FieldErr(err), log.FieldKey(key))
	}
	return c
}
