package egrpc

import (
	"time"

	"miopkg/conf"
	"miopkg/errors"
	ig "miopkg/grpc"
	"miopkg/log"
	"miopkg/util/xtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

// Config ...
type Config struct {
	Name                         string        // config's name
	Address                      string        // 连接地址，直连为127.0.0.1:9001，服务发现为etcd:///appname
	BalancerName                 string        // 负载均衡方式，默认round robin
	OnFail                       string        // 失败后的处理方式，panic | error
	DialTimeout                  time.Duration // 连接超时，默认3s
	ReadTimeout                  time.Duration // 读超时，默认1s
	SlowLogThreshold             time.Duration // 慢日志记录的阈值，默认600ms
	EnableBlock                  bool          // 是否开启阻塞，默认开启
	EnableOfficialGrpcLog        bool          // 是否开启官方grpc日志，默认关闭
	EnableWithInsecure           bool          // 是否开启非安全传输，默认开启
	EnableMetricInterceptor      bool          // 是否开启监控，默认开启
	EnableTraceInterceptor       bool          // 是否开启链路追踪，默认关闭
	EnableAppNameInterceptor     bool          // 是否开启传递应用名，默认开启
	EnableTimeoutInterceptor     bool          // 是否开启超时传递，默认开启
	EnableAccessInterceptor      bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorReq   bool          // 是否开启记录请求参数，默认不开启
	EnableAccessInterceptorRes   bool          // 是否开启记录响应参数，默认不开启
	EnableCPUUsage               bool          // 是否开启CPU利用率，默认开启
	EnableServiceConfig          bool          // 是否开启服务配置，默认关闭
	EnableFailOnNonTempDialError bool

	keepAlive   *keepalive.ClientParameters
	dialOptions []grpc.DialOption

	logger                 *log.Logger
	Direct                 bool
	OnDialError            string // panic | error
	Debug                  bool
	AccessInterceptorLevel string
}

// DefaultConfig defines grpc client default configuration
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		dialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
		BalancerName:                 roundrobin.Name,
		OnFail:                       "panic",
		DialTimeout:                  time.Second * 3,
		ReadTimeout:                  xtime.Duration("1s"),
		SlowLogThreshold:             xtime.Duration("600ms"),
		EnableBlock:                  true,
		EnableTraceInterceptor:       false,
		EnableWithInsecure:           true,
		EnableAppNameInterceptor:     true,
		EnableTimeoutInterceptor:     true,
		EnableMetricInterceptor:      true,
		EnableFailOnNonTempDialError: true,
		EnableAccessInterceptor:      false,
		EnableAccessInterceptorReq:   false,
		EnableAccessInterceptorRes:   false,
		EnableServiceConfig:          false,
		EnableCPUUsage:               true,
		logger:                       log.MioLogger.With(log.FieldMod(errors.ModClientGrpc)),
		AccessInterceptorLevel:       "info",
	}
}

// Option 可选项
type Option func(c *Config)

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig("mio.client." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("client grpc parse config panic", log.FieldErrKind(errors.ErrKindUnmarshalConfigErr), log.FieldErr(err), log.FieldKey(key), log.FieldValueAny(config))
	}
	return config
}

// Build 构建组件
func (c *Config) Build(options ...Option) *grpc.ClientConn {
	// 最先执行trace
	if c.EnableTraceInterceptor {
		options = append(options,
			WithDialOption(grpc.WithChainUnaryInterceptor(c.traceUnaryClientInterceptor())),
		)
	}

	// 其次执行，自定义header头，这样才能赋值到ctx里
	options = append(options, WithDialOption(grpc.WithChainUnaryInterceptor(customHeader(ig.CustomContextKeys()))))

	// 默认日志
	options = append(options, WithDialOption(grpc.WithChainUnaryInterceptor(c.loggerUnaryClientInterceptor())))

	if c.Debug {
		options = append(options, WithDialOption(grpc.WithChainUnaryInterceptor(c.debugUnaryClientInterceptor(c.Address))))
	}

	if c.EnableAppNameInterceptor {
		options = append(options, WithDialOption(grpc.WithChainUnaryInterceptor(c.defaultUnaryClientInterceptor())))
		options = append(options, WithDialOption(grpc.WithChainStreamInterceptor(c.defaultStreamClientInterceptor())))
	}

	if c.EnableTimeoutInterceptor {
		options = append(options, WithDialOption(grpc.WithChainUnaryInterceptor(c.timeoutUnaryClientInterceptor())))
	}

	// 定位到bug,关闭就没有bug.inconsistent label cardinality: expected 5 label values but got 4 in []string
	if c.EnableMetricInterceptor {
		options = append(options,
			WithDialOption(grpc.WithChainUnaryInterceptor(c.metricUnaryClientInterceptor(c.Name))),
		)
	}

	for _, option := range options {
		option(c)
	}

	return newGRPCClient(c)
}
