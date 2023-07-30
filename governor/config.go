package governor

import (
	"fmt"

	"miopkg/conf"
	"miopkg/log"
	"miopkg/util/xnet"
)

// ModName ..
const ModName = "govern"

// Config ...
type Config struct {
	Host    string
	Port    int
	Network string `json:"network" toml:"network"`
	logger  *log.Logger
	Enable  bool

	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string
}

// StdConfig represents Standard gRPC Server config
// which will parse config by conf package,
// panic if no config key found in conf
func StdConfig(name string) *Config {
	return RawConfig("mio.server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if conf.Get(key) == nil {
		return config
	}
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("govern server parse config panic",
			log.FieldErr(err), log.FieldKey(key),
			log.FieldValueAny(config),
		)
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	host, port, err := xnet.GetLocalMainIP()
	if err != nil {
		host = "localhost"
	}
	return &Config{
		Enable:  true,
		Host:    host,
		Network: "tcp4",
		Port:    port,
		logger:  log.MioLogger.With(log.FieldMod(ModName)),
	}
}

// Build governor server
func (config *Config) Build() *Server {
	return newServer(config)
}

// Address print host:port
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
