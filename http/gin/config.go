package gin

import (
	"crypto/tls"
	"embed"
	"fmt"
	"miopkg/conf"
	ierrors "miopkg/errors"
	"miopkg/flag"
	"miopkg/log"
	"miopkg/trace"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ModName ..
const ModName = "server.gin"

// Config HTTP config
type Config struct {
	Host          string // IP地址，默认从flag的host读取
	Port          int    // PORT端口，默认9001
	Deployment    string
	Mode          string // gin的模式，默认是release模式
	DisableMetric bool   // 是否开启监控，默认开启
	DisableTrace  bool   // 是否开启链路追踪，默认开启

	ServiceAddress string // ServiceAddress service address in registry info, default to 'Host:Port'

	SlowQueryThresholdInMilli int64 // 服务慢日志，默认500ms;可以设置为time.Duration
	// SlowLogThreshold           time.Duration // 服务慢日志，默认500ms
	logger *log.Logger

	EnableLocalMainIP bool // 自动获取ip地址

	EnableAccessInterceptor    bool          // 是否开启，记录请求数据
	EnableAccessInterceptorReq bool          // 是否开启记录请求参数，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
	EnableTrustedCustomHeader  bool          // 是否开启自定义header头，记录数据往链路后传递，默认不开启
	EnableSentinel             bool          // 是否开启限流，默认不开启
	WebsocketHandshakeTimeout  time.Duration // 握手时间
	WebsocketReadBufferSize    int
	WebsocketWriteBufferSize   int
	EnableWebsocketCompression bool     // 是否开通压缩
	EnableWebsocketCheckOrigin bool     // 是否支持跨域
	EnableTLS                  bool     // 是否进入 https 模式
	TLSCertFile                string   // https 证书
	TLSKeyFile                 string   // https 私钥
	TLSClientAuth              string   // https 客户端认证方式默认为 NoClientCert(NoClientCert,RequestClientCert,RequireAnyClientCert,VerifyClientCertIfGiven,RequireAndVerifyClientCert)
	TLSClientCAs               []string // https client的ca，当需要双向认证的时候指定可以倒入自签证书
	TrustedPlatform            string   // 需要用户换成自己的CDN名字，获取客户端IP地址
	EmbedPath                  string   // 嵌入embed path数据
	embedFs                    embed.FS // 需要在build时候注入embed.Fs
	TLSSessionCache            tls.ClientSessionCache
	blockFallback              func(*gin.Context)
	resourceExtract            func(*gin.Context) string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                       flag.String("host"),
		Port:                       9091,
		Mode:                       gin.ReleaseMode,
		SlowQueryThresholdInMilli:  500, // 500ms
		logger:                     log.MioLogger.With(log.FieldMod(ModName)),
		EnableWebsocketCheckOrigin: false,
		TrustedPlatform:            "",
	}
}

// / 1 从这里开始调用 server := xgin.StdConfig("http").Build()
// StdConfig mio Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig("mio.server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", log.FieldErrKind(ierrors.ErrKindUnmarshalConfigErr), log.FieldErr(err), log.FieldKey(key), log.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *log.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

// Build create server instance, then initialize it with necessary interceptor
// / 2调用newServer生成server
// / 根据配置传入中间件,use方法是gin自带的中间件使用方法
func (config *Config) Build(options ...Option) *Server {
	for _, option := range options {
		option(config)
	}

	server := newServer(config)
	server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))

	if !config.DisableMetric {
		server.Use(metricServerInterceptor())
	}

	if !config.DisableTrace && trace.IsGlobalTracerRegistered() {
		server.Use(traceServerInterceptor())
	}
	return server
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// ClientAuthType 客户端auth类型
func (config *Config) ClientAuthType() tls.ClientAuthType {
	switch config.TLSClientAuth {
	case "NoClientCert":
		return tls.NoClientCert
	case "RequestClientCert":
		return tls.RequestClientCert
	case "RequireAnyClientCert":
		return tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		return tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}
