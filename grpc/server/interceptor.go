package grpcsvr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"miopkg/errors"
	ig "miopkg/grpc"
	"miopkg/log"
	"miopkg/trace"
	"miopkg/util/cpu"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	grpcmd "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"miopkg/metric"

	ocodes "go.opentelemetry.io/otel/codes"
	otrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func prometheusUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	resp, err := handler(ctx, req)
	statusInfo := errors.Convert(err)
	metric.ServerHandleHistogram.Observe(time.Since(startTime).Seconds(), metric.TypeGRPCUnary, info.FullMethod, extractAID(ctx))
	metric.ServerHandleCounter.Inc(metric.TypeGRPCUnary, info.FullMethod, extractAID(ctx), statusInfo.Message(), http.StatusText(errors.GrpcToHTTPStatusCode(statusInfo.Code())))
	return resp, err
}

func prometheusStreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	startTime := time.Now()
	err := handler(srv, ss)
	statusInfo := errors.Convert(err)
	metric.ServerHandleHistogram.Observe(time.Since(startTime).Seconds(), metric.TypeGRPCStream, info.FullMethod, extractAID(ss.Context()))
	metric.ServerHandleCounter.Inc(metric.TypeGRPCStream, info.FullMethod, extractAID(ss.Context()), statusInfo.Message())
	return err
}

func traceUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := trace.NewTracer(otrace.SpanKindServer)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
		md, ok := grpcmd.FromIncomingContext(ctx)
		if !ok {
			md = grpcmd.New(nil)
		}
		// Deprecated 该方法会在v0.9.0移除
		trace.CompatibleExtractGrpcTraceID(md)
		ctx, span := tracer.Start(ctx, info.FullMethod, ig.GrpcHeaderCarrier(md))
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", info.FullMethod),
			// attribute.String("net.peer.name", getPeerName(ctx)),
			// attribute.String("net.peer.ip", getPeerIP(ctx)),
			trace.TagSpanKind("server.unary"),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(attribute.Key("rpc.grpc.status_code").Int64(int64(e.Code)))
				}
				span.SetStatus(ocodes.Error, err.Error())
			} else {
				span.SetStatus(ocodes.Ok, "OK")
			}
			span.End()
		}()
		return handler(ctx, req)
	}
}

type contextedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context ...
func (css contextedServerStream) Context() context.Context {
	return css.ctx
}
func traceStreamServerInterceptor() grpc.StreamServerInterceptor {
	tracer := trace.NewTracer(otrace.SpanKindServer)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := grpcmd.FromIncomingContext(ss.Context())
		if !ok {
			md = grpcmd.New(nil)
		}
		// Deprecated 该方法会在v0.9.0移除
		trace.CompatibleExtractGrpcTraceID(md)
		ctx, span := tracer.Start(ss.Context(), info.FullMethod, ig.GrpcHeaderCarrier(md))
		span.SetAttributes(
			trace.TagComponent("grpc"),
			trace.TagSpanKind("server.stream"),
		)
		defer span.End()
		return handler(srv, contextedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func extractAID(ctx context.Context) string {
	if md, ok := grpcmd.FromIncomingContext(ctx); ok {
		return strings.Join(md.Get("aid"), ",")
	}
	return "unknown"
}

func defaultStreamServerInterceptor(logger *log.Logger, c *Config) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var beg = time.Now()
		var fields = make([]log.Field, 0, 8)
		var event = "normal"
		defer func() {
			if c.SlowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > c.SlowQueryThresholdInMilli {
					event = "slow"
				}
			}

			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				fields = append(fields, log.FieldStack(stack))
				event = "recover"
			}

			fields = append(fields,
				log.Any("grpc interceptor type", "unary"),
				log.FieldMethod(info.FullMethod),
				log.FieldCost(time.Since(beg)),
				log.FieldEvent(event),
			)

			for key, val := range getPeer(stream.Context()) {
				fields = append(fields, log.Any(key, val))
			}

			if err != nil {
				fields = append(fields, zap.String("err", err.Error()))
				logger.Error("access", fields...)
				return
			}

			if c.EnableAccessLog {
				logger.Info("access", fields...)
			}
		}()
		return handler(srv, stream)
	}
}

func defaultUnaryServerInterceptor(logger *log.Logger, c *Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 默认过滤掉该探活日志
		if c.EnableSkipHealthLog && info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		var beg = time.Now()
		// 为了性能考虑，如果要加日志字段，需要改变slice大小
		loggerKeys := ig.CustomContextKeys()
		var fields = make([]log.Field, 0, 20+ig.CustomContextKeysLength())
		var event = "normal"
		// 必须在defer外层，因为要赋值，替换ctx
		for _, key := range loggerKeys {
			if value := ig.GrpcHeaderValue(ctx, key); value != "" {
				ctx = ig.WithValue(ctx, key, value)
			}
		}
		// 此处必须使用defer来recover handler内部可能出现的panic
		defer func() {
			if c.SlowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > c.SlowQueryThresholdInMilli {
					event = "slow"
				}
			}
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}

				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				fields = append(fields, log.FieldStack(stack))
				event = "recover"
			}
			// TODO: log更多信息 ,参考other
			// spbStatus := errors.Convert(err)
			// httpStatusCode := errors.GrpcToHTTPStatusCode(spbStatus.Code())

			fields = append(fields,
				log.Any("grpc interceptor type", "unary"),
				log.FieldMethod(info.FullMethod),
				log.FieldCost(time.Since(beg)),
				log.FieldEvent(event),
			)

			for key, val := range getPeer(ctx) {
				fields = append(fields, log.Any(key, val))
			}

			if err != nil {
				fields = append(fields, zap.String("err", err.Error()))
				logger.Error("access", fields...)
				return
			}

			if c.EnableAccessLog {
				logger.Info("access", fields...)
			}
		}()
		// 用于p2c和限流、熔断
		if enableCPUUsage(ctx) {
			var stat = cpu.Stat{}
			cpu.ReadStat(&stat)
			if stat.Usage > 0 {
				// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
				header := grpcmd.Pairs("cpu-usage", strconv.Itoa(int(stat.Usage)))
				err = grpc.SetHeader(ctx, header)
				if err != nil {
					c.logger.Error("set header error", log.FieldErr(err))
				}
			}
		}
		return handler(ctx, req)
	}
}

// enableCPUUsage 是否开启cpu利用率
func enableCPUUsage(ctx context.Context) bool {
	return ig.GrpcHeaderValue(ctx, "enable-cpu-usage") == "true"
}

// getPeerName 获取对端应用名称
func getPeerName(ctx context.Context) string {
	return ig.GrpcHeaderValue(ctx, "app")
}

// getPeerIP 获取对端ip
func getPeerIP(ctx context.Context) string {
	clientIP := ig.GrpcHeaderValue(ctx, "client-ip")
	if clientIP != "" {
		return clientIP
	}

	// 从grpc里取对端ip
	pr, ok2 := peer.FromContext(ctx)
	if !ok2 {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

func getPeer(ctx context.Context) map[string]string {
	var peerMeta = make(map[string]string)
	if md, ok := grpcmd.FromIncomingContext(ctx); ok {
		if val, ok := md["aid"]; ok {
			peerMeta["aid"] = strings.Join(val, ";")
		}
		var clientIP string
		if val, ok := md["client-ip"]; ok {
			clientIP = strings.Join(val, ";")
		} else {
			ip, err := getClientIP(ctx)
			if err == nil {
				clientIP = ip
			}
		}
		peerMeta["clientIP"] = clientIP
		if val, ok := md["client-host"]; ok {
			peerMeta["host"] = strings.Join(val, ";")
		}
	}
	return peerMeta

}
