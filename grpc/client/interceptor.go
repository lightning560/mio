package egrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"miopkg/env"
	"miopkg/errors"
	ig "miopkg/grpc"
	ilog "miopkg/log"
	"miopkg/metric"
	"miopkg/trace"

	"miopkg/util/xstring"

	"github.com/fatih/color"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otrace "go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// metric统计
// metricUnaryClientInterceptor returns grpc unary request metrics collector interceptor
// FIXME: 少一个name造成无法对应libels,就会报错
func (c *Config) metricUnaryClientInterceptor(name string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		statusInfo := errors.Convert(err)
		metric.ClientHandleCounter.Inc(metric.TypeGRPCUnary, name, method, cc.Target(), statusInfo.Message())
		metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeGRPCUnary, name, method, cc.Target())
		return err
	}
}

// debugUnaryClientInterceptor returns grpc unary request request and response details interceptor
func (c *Config) debugUnaryClientInterceptor(addr string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var p peer.Peer
		beg := time.Now()
		prefix := fmt.Sprintf("[%s]", addr)
		if remote, ok := peer.FromContext(ctx); ok && remote.Addr != nil {
			prefix = prefix + "(" + remote.Addr.String() + ")"
		}
		fmt.Printf("%-50s[%s] => %s\n", color.GreenString(prefix), time.Now().Format("04:05.000"), color.GreenString("Send: "+method+" | "+xstring.Json(req)))
		err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Peer(&p))...)
		cost := time.Since(beg)
		if err != nil {
			log.Println("grpc.response", ilog.MakeReqResError(0, c.Name, c.Address, cost, method+" | "+fmt.Sprintf("%v", req), err.Error()))
		} else {
			log.Println("grpc.response", ilog.MakeReqResInfo(0, c.Name, c.Address, cost, method+" | "+fmt.Sprintf("%v", req), reply))
		}
		return err
	}
}

// traceUnaryClientInterceptor returns grpc unary opentracing interceptor
func (c *Config) traceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := trace.NewTracer(otrace.SpanKindClient)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		ctx, span := tracer.Start(ctx, method, ig.GrpcHeaderCarrier(md))
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("net.peer.name", c.Address),
		)
		// 因为这里最先执行trace，所以这里，直接new出来metadata
		ctx = metadata.NewOutgoingContext(ctx, md)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(attribute.Key("rpc.grpc.status_code").Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// defaultUnaryClientInterceptor returns interceptor inject app name
func (c *Config) defaultUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
		ctx = metadata.AppendToOutgoingContext(ctx, "app", env.Name())
		if c.EnableCPUUsage {
			ctx = metadata.AppendToOutgoingContext(ctx, "enable-cpu-usage", "true")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Config) defaultStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
		ctx = metadata.AppendToOutgoingContext(ctx, "app", env.Name())
		if c.EnableCPUUsage {
			ctx = metadata.AppendToOutgoingContext(ctx, "enable-cpu-usage", "true")
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// timeoutUnaryClientInterceptor settings timeout
func (c *Config) timeoutUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		now := time.Now()
		// 若无自定义超时设置，默认设置超时
		_, ok := ctx.Deadline()
		if !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.ReadTimeout)
			defer cancel()
		}
		// return invoker(ctx, method, req, reply, cc, opts...)

		err := invoker(ctx, method, req, reply, cc, opts...)
		du := time.Since(now)
		remoteIP := "unknown"
		if remote, ok := peer.FromContext(ctx); ok && remote.Addr != nil {
			remoteIP = remote.Addr.String()
		}

		if c.SlowLogThreshold > time.Duration(0) && du > c.SlowLogThreshold {
			c.logger.Error("slow",
				ilog.FieldErr(errors.New(4324234, "errSlowCommand", "grpc client slow")),
				ilog.FieldMethod(method),
				ilog.FieldName(cc.Target()),
				ilog.FieldCost(du),
				ilog.FieldAddr(remoteIP),
			)
		}
		return err
	}
}

// loggerUnaryClientInterceptor returns log interceptor for logging
func (c *Config) loggerUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		loggerKeys := ig.CustomContextKeys()
		var fields = make([]ilog.Field, 0, 20+ig.CustomContextKeysLength())

		for _, key := range loggerKeys {
			if value := ig.ContextValue(ctx, key); value != "" {
				fields = append(fields, ilog.FieldCustomKeyValue(key, value))
				// 替换context
				ctx = metadata.AppendToOutgoingContext(ctx, key, value)
			}
		}

		err := invoker(ctx, method, req, res, cc, opts...)
		cost := time.Since(beg)
		spbStatus := errors.Convert(err)
		httpStatusCode := errors.GrpcToHTTPStatusCode(spbStatus.Code())

		fields = append(fields,
			ilog.FieldType("unary"),
			ilog.FieldCode(int32(spbStatus.Code())),
			ilog.FieldUniformCode(int32(httpStatusCode)),
			ilog.FieldDescription(spbStatus.Message()),
			ilog.FieldMethod(method),
			ilog.FieldCost(cost),
			ilog.FieldName(cc.Target()),
		)

		// 开启了链路，那么就记录链路id
		if c.EnableTraceInterceptor && trace.IsGlobalTracerRegistered() {
			fields = append(fields, ilog.FieldTid(trace.ExtractTraceID(ctx)))
		}

		if c.EnableAccessInterceptorReq {
			fields = append(fields, ilog.Any("req", json.RawMessage(xstring.Json(req))))
		}
		if c.EnableAccessInterceptorRes {
			fields = append(fields, ilog.Any("res", json.RawMessage(xstring.Json(res))))
		}

		if c.SlowLogThreshold > time.Duration(0) && cost > c.SlowLogThreshold {
			c.logger.Warn("slow", fields...)
		}

		if err != nil {
			fields = append(fields, ilog.FieldEvent("error"), ilog.FieldErr(err))
			// 只记录系统级别错误
			if httpStatusCode >= http.StatusInternalServerError {
				// 只记录系统级别错误
				c.logger.Error("access", fields...)
				return err
			}
			// 业务报错只做warning
			c.logger.Warn("access", fields...)
			return err
		}

		if c.EnableAccessInterceptor {
			fields = append(fields, ilog.FieldEvent("normal"))
			c.logger.Info("access", fields...)
		}
		return nil
	}
}

// customHeader 自定义header头
func customHeader(logExtraKeys []string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		for _, key := range logExtraKeys {
			if value := ig.GrpcHeaderValue(ctx, key); value != "" {
				ctx = ig.WithValue(ctx, key, value)
			}
		}
		return invoker(ctx, method, req, res, cc, opts...)
	}
}
