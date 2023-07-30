package etcdv3

import (
	"time"

	"miopkg/errors"
	"miopkg/registry"

	"miopkg/client/etcdv3"
	"miopkg/conf"
	"miopkg/log"
)

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig("mio.registry." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	// 解析最外层配置
	if err := conf.UnmarshalKey(key, &config); err != nil {
		log.Panic("unmarshal key", log.FieldMod("registry.etcd"), log.FieldErrKind(errors.ErrKindUnmarshalConfigErr), log.FieldErr(err), log.String("key", key), log.Any("config", config))
	}
	// 解析嵌套配置
	if err := conf.UnmarshalKey(key, &config.Config); err != nil {
		log.Panic("unmarshal key", log.FieldMod("registry.etcd"), log.FieldErrKind(errors.ErrKindUnmarshalConfigErr), log.FieldErr(err), log.String("key", key), log.Any("config", config))
	}
	return config
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Config:      etcdv3.DefaultConfig(),
		ReadTimeout: time.Second * 3,
		Prefix:      "mio",
		logger:      log.MioLogger,
		ServiceTTL:  0,
	}
}

// Config ...
type Config struct {
	*etcdv3.Config
	ReadTimeout time.Duration
	ConfigKey   string
	Prefix      string
	ServiceTTL  time.Duration
	logger      *log.Logger
}

// Build ...
func (config Config) Build() (registry.Registry, error) {
	if config.ConfigKey != "" {
		config.Config = etcdv3.RawConfig(config.ConfigKey)
	}
	return newETCDRegistry(&config)
}

func (config Config) MustBuild() registry.Registry {
	reg, err := config.Build()
	if err != nil {
		log.Panicf("build registry failed: %v", err)
	}
	return reg
}
