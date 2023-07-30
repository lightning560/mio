package egrpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// WithAddr setting grpc server address
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Address = addr
	}
}

// WithOnFail setting failing mode
func WithOnFail(onFail string) Option {
	return func(c *Config) {
		c.OnFail = onFail
	}
}

// WithBalancerName setting grpc load balancer name
func WithBalancerName(balancerName string) Option {
	return func(c *Config) {
		c.BalancerName = balancerName
	}
}

// WithDialTimeout setting grpc dial timeout
func WithDialTimeout(t time.Duration) Option {
	return func(c *Config) {
		c.DialTimeout = t
	}
}

// WithReadTimeout setting grpc read timeout
func WithReadTimeout(t time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = t
	}
}

// WithDebug setting if enable debug mode
func WithDebug(enableDebug bool) Option {
	return func(c *Config) {
		// for version compatibility
	}
}

// WithDialOption setting grpc dial options
func WithDialOption(opts ...grpc.DialOption) Option {
	return func(c *Config) {
		if c.dialOptions == nil {
			c.dialOptions = make([]grpc.DialOption, 0)
		}
		c.dialOptions = append(c.dialOptions, opts...)
	}
}

// WithEnableAccessInterceptor 开启日志记录
func WithEnableAccessInterceptor(enableAccessInterceptor bool) Option {
	return func(c *Config) {
		c.EnableAccessInterceptor = enableAccessInterceptor
	}
}

// WithEnableAccessInterceptorReq 开启日志请求参数
func WithEnableAccessInterceptorReq(enableAccessInterceptorReq bool) Option {
	return func(c *Config) {
		c.EnableAccessInterceptorReq = enableAccessInterceptorReq
	}
}

// WithEnableAccessInterceptorRes 开启日志响应记录
func WithEnableAccessInterceptorRes(enableAccessInterceptorRes bool) Option {
	return func(c *Config) {
		c.EnableAccessInterceptorRes = enableAccessInterceptorRes
	}
}

// WithBufnetServerListener 写入bufnet listener
func WithBufnetServerListener(svc net.Listener) Option {
	return WithDialOption(grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return svc.(*bufconn.Listener).Dial()
	}))
}

// WithName name
func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}
