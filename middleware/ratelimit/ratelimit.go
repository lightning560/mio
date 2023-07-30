package ratelimit

import (
	"context"
	"miopkg/errors"
	"miopkg/middleware"

	"miopkg/middleware/ratelimit/bbr"
)

// Option is ratelimit option.
type Option func(*options)

// WithLimiter set Limiter implementation,
// default is bbr limiter
func WithLimiter(limiter bbr.Limiter) Option {
	return func(o *options) {
		o.limiter = limiter
	}
}

type options struct {
	limiter bbr.Limiter
}

// Server ratelimiter middleware
func Server(opts ...Option) middleware.Middleware {
	options := &options{
		limiter: bbr.NewLimiter(),
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			done, e := options.limiter.Allow()
			if e != nil {
				// rejected
				return nil, errors.ErrLimitExceed
			}
			// allowed
			reply, err = handler(ctx, req)
			done(bbr.DoneInfo{Err: err})
			return
		}
	}
}
