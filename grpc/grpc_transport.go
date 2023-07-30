package transport

import (
	"context"
	"strings"

	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"
)

// GrpcHeaderCarrier ...
type GrpcHeaderCarrier metadata.MD

// Get returns the value associated with the passed key.
func (mc GrpcHeaderCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set stores the key-value pair.
func (mc GrpcHeaderCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (mc GrpcHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}

// GrpcHeaderValue 获取context value
func GrpcHeaderValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	// 小写
	return strings.Join(md.Get(key), ";")
}

// ContextValue gRPC日志获取context value
func ContextValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	return cast.ToString(Value(ctx, key))
}
