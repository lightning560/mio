package grpcsvr

import (
	"fmt"
	"time"

	"miopkg/flag"

	"miopkg/errors"
	"miopkg/log"
	"miopkg/util/constant"

	"miopkg/conf"

	"google.golang.org/grpc"
)

// Config ...
type Config struct {
	Name       string `json:"name"`
	Host       string `json:"host"`       // IP地址，默认0.0.0.0
	Port       int    `json:"port"`       // Port端口，默认9002
	Deployment string `json:"deployment"` // 部署区域

	// Network network type, tcp4 by default
	Network string `json:"network" toml:"network"` // 网络类型，默认tcp4
	// EnableAccessLog enable Access Interceptor, true by default
	EnableAccessLog bool // 是否开启，记录请求数据,默认开启
	// DisableTrace disbale Trace Interceptor, false by default
	EnableTrace bool // 是否开启链路追踪，默认关闭
	// DisableMetric disable Metric Interceptor, false by default
	EnableMetric bool // 是否开启官方grpc日志，默认关闭
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64 // 服务慢日志，默认500ms

	ServiceAddress string // ServiceAddress service address in registry info, default to 'Host:Port'

	EnableTLS bool // EnableTLS

	CaFile string // CaFile

	CertFile string // CertFile

	PrivateFile string // PrivateFile

	Labels map[string]string `json:"labels"`

	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor

	logger                     *log.Logger
	EnableOfficialGrpcLog      bool // 是否开启官方grpc日志，默认关闭
	EnableSkipHealthLog        bool // 是否屏蔽探活日志，默认false
	EnableAccessInterceptorReq bool // 是否开启记录请求参数，默认不开启
	EnableAccessInterceptorRes bool // 是否开启记录响应参数，默认不开启
	EnableLocalMainIP          bool // 自动获取ip地址
	timeout                    time.Duration
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
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("grpc server parse config panic",
			log.FieldErrKind(errors.ErrKindUnmarshalConfigErr),
			log.FieldErr(err), log.FieldKey(key),
			log.FieldValueAny(config),
		)
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      flag.String("host"),
		Port:                      9092,
		Deployment:                constant.DefaultDeployment,
		EnableAccessLog:           true,
		EnableMetric:              false,
		EnableTrace:               false,
		EnableTLS:                 false,
		SlowQueryThresholdInMilli: 500,
		logger:                    log.MioLogger.With(log.FieldMod("server.grpc")),

		EnableSkipHealthLog:        false,
		EnableAccessInterceptorReq: false,
		EnableAccessInterceptorRes: false,
		serverOptions:              []grpc.ServerOption{},
		streamInterceptors:         []grpc.StreamServerInterceptor{},
		unaryInterceptors:          []grpc.UnaryServerInterceptor{},
	}
}

// Option 可选项
type Option func(c *Config)

// WithServerOption inject server option to grpc server
// User should not inject interceptor option, which is recommend by WithStreamInterceptor
// and WithUnaryInterceptor
func (config *Config) WithServerOption(options ...grpc.ServerOption) Option {
	return func(c *Config) {
		if config.serverOptions == nil {
			config.serverOptions = make([]grpc.ServerOption, 0)
		}
		config.serverOptions = append(config.serverOptions, options...)
	}
}

// WithStreamInterceptor inject stream interceptors to server option
func (config *Config) WithStreamInterceptor(intes ...grpc.StreamServerInterceptor) Option {
	return func(c *Config) {
		if config.streamInterceptors == nil {
			config.streamInterceptors = make([]grpc.StreamServerInterceptor, 0)
		}

		config.streamInterceptors = append(config.streamInterceptors, intes...)
	}
}

// WithUnaryInterceptor inject unary interceptors to server option
func (config *Config) WithUnaryInterceptor(intes ...grpc.UnaryServerInterceptor) Option {
	return func(c *Config) {
		if config.unaryInterceptors == nil {
			config.unaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
		}

		config.unaryInterceptors = append(config.unaryInterceptors, intes...)
	}
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		log.Panicf("build xgrpc server: %v", err)
	}
	return server
}

// Build ...
// / 这里加载interceptor,并没有执行，仅仅是处理配置。 最后调用的newServer才是具体执行的.
func (config *Config) Build(options ...Option) (*Server, error) {
	var streamInterceptors []grpc.StreamServerInterceptor
	var unaryInterceptors []grpc.UnaryServerInterceptor
	// trace 必须在最外层，否则无法取到trace信息，传递到其他中间件
	// TODO: 可以改为option的设计模式
	if config.EnableTrace {
		unaryInterceptors = []grpc.UnaryServerInterceptor{traceUnaryServerInterceptor(), defaultUnaryServerInterceptor(config.logger, config)}
		streamInterceptors = []grpc.StreamServerInterceptor{traceStreamServerInterceptor(), defaultStreamServerInterceptor(config.logger, config)}
	} else {
		unaryInterceptors = []grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config)}
		streamInterceptors = []grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config)}
	}
	// 使用option的方式，然后用chainxxx方法执行
	if config.EnableMetric {
		options = append(options, config.WithUnaryInterceptor(prometheusUnaryServerInterceptor))
		options = append(options, config.WithStreamInterceptor(prometheusStreamServerInterceptor))
	}

	for _, option := range options {
		option(config)
	}

	streamInterceptors = append(
		streamInterceptors,
		config.streamInterceptors...,
	)

	unaryInterceptors = append(
		unaryInterceptors,
		config.unaryInterceptors...,
	)
	config.serverOptions = append(config.serverOptions,
		grpc.ChainStreamInterceptor(streamInterceptors...),
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	)

	return newServer(config)
}

// WithLogger ...
func (config *Config) WithLogger(logger *log.Logger) *Config {
	config.logger = logger
	return config
}

// Address ...
func (config Config) Address() string {
	// 如果是unix，那么启动方式为unix domain socket，host填写file
	if config.Network == "unix" {
		return config.Host
	}
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
