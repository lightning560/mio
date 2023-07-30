package gin

import (
	"crypto/tls"
	"embed"

	"miopkg/log"

	"github.com/gin-gonic/gin"
)

type Option func(c *Config)

// WithSentinelResourceExtractor 资源命名方式
func WithSentinelResourceExtractor(fn func(*gin.Context) string) Option {
	return func(c *Config) {
		c.resourceExtract = fn
	}
}

// WithSentinelBlockFallback 限流后的返回数据
func WithSentinelBlockFallback(fn func(*gin.Context)) Option {
	return func(c *Config) {
		c.blockFallback = fn
	}
}

// WithTLSSessionCache 限流后的返回数据
func WithTLSSessionCache(tsc tls.ClientSessionCache) Option {
	return func(c *Config) {
		c.TLSSessionCache = tsc
	}
}

// WithTrustedPlatform 信任的Header头，获取客户端IP地址
func WithTrustedPlatform(trustedPlatform string) Option {
	return func(c *Config) {
		c.TrustedPlatform = trustedPlatform
	}
}

// WithLogger 信任的Header头，获取客户端IP地址
func WithLogger(logger *log.Logger) Option {
	return func(c *Config) {
		c.logger = logger
	}
}

// WithEmbedFs 设置embed fs
func WithEmbedFs(fs embed.FS) Option {
	return func(c *Config) {
		c.embedFs = fs
	}
}
