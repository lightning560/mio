package etcdv3

import (
	"time"

	"miopkg/conf"
	"miopkg/errors"
	"miopkg/flag"
	"miopkg/log"
	"miopkg/util/constant"
	"miopkg/util/xtime"
)

var ConfigPrefix = constant.ConfigPrefix + ".etcdv3"

// Config ...
type (
	Config struct {
		Endpoints        []string      `json:"endpoints"` // 地址
		CertFile         string        `json:"certFile"`  // cert file
		KeyFile          string        `json:"keyFile"`   // key file
		CaCert           string        `json:"caCert"`    // ca cert
		BasicAuth        bool          `json:"basicAuth"`
		UserName         string        `json:"userName"`          // 用户名
		Password         string        `json:"-"`                 // 密码
		ConnectTimeout   time.Duration `json:"connectTimeout"`    // 连接超时时间
		Secure           bool          `json:"secure"`            // 是否开启安全
		AutoSyncInterval time.Duration `json:"autoAsyncInterval"` // 自动同步member list的间隔
		TTL              int           // 单位：s
		logger           *log.Logger

		EnableBlock                  bool // 是否开启阻塞，默认开启
		EnableFailOnNonTempDialError bool // 是否开启gRPC连接的错误信息
	}
)

func (config *Config) BindFlags(fs *flag.FlagSet) {
	fs.BoolVar(&config.Secure, "insecure-etcd", true, "--insecure-etcd=true")
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BasicAuth:                    false,
		ConnectTimeout:               xtime.Duration("5s"),
		Secure:                       false,
		logger:                       log.MioLogger.With(log.FieldMod("client.etcd")),
		EnableBlock:                  true,
		EnableFailOnNonTempDialError: true,
	}
}

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(ConfigPrefix + "." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, config); err != nil {
		config.logger.Panic("client etcd parse config panic", log.FieldErrKind(errors.ErrKindUnmarshalConfigErr), log.FieldErr(err), log.FieldKey(key), log.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *log.Logger) *Config {
	config.logger = logger
	return config
}

// Build ...
func (config *Config) Build() (*Client, error) {
	return newClient(config)
}

func (config *Config) MustBuild() *Client {
	client, err := config.Build()
	if err != nil {
		log.Panicf("build etcd client failed: %v", err)
	}
	return client
}
