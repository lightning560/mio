package egrpc

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"miopkg/errors"
	"miopkg/log"
)

// PackageName 设置包名
const PackageName = "client.grpc"

func newGRPCClient(config *Config) *grpc.ClientConn {
	if config.EnableOfficialGrpcLog {
		// grpc框架日志，因为官方grpc日志是单例，所以这里要处理下
		grpclog.SetLoggerV2(zapgrpc.NewLogger(grpcLogBuild().ZapLogger()))
	}
	var ctx = context.Background()
	var dialOptions = config.dialOptions
	logger := config.logger.With(
		log.FieldMod("client.grpc"),
		log.FieldAddr(config.Address),
	)
	// 默认配置使用block
	if config.EnableBlock {
		if config.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
			defer cancel()
		}

		dialOptions = append(dialOptions, grpc.WithBlock())
	}

	if config.EnableWithInsecure {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	if config.keepAlive != nil {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(*config.keepAlive))
	}

	// 因为默认是开启这个配置
	// 并且开启后，在grpc 1.40以上会导致dns多一次解析txt内容（目测是为了做grpc的load balance策略，但我们实际上不会用到）
	// 因为这个service config dns域名通常是没有设置dns解析，所以会跳过k8s的dns，穿透到上一级的dns，而如果dns配置有问题或者不存在，那么会查询非常长的时间（通常在20s或者更长）
	// 那么为false的时候，禁用他，可以加快我们的启动时间或者提升我们的性能
	if !config.EnableServiceConfig {
		dialOptions = append(dialOptions, grpc.WithDisableServiceConfig())
	}
	//使用的grpc自带的balance。还有backoff.不过两个都Deprecated了
	dialOptions = append(dialOptions, grpc.WithBalancerName(config.BalancerName)) //nolint
	dialOptions = append(dialOptions, grpc.FailOnNonTempDialError(config.EnableFailOnNonTempDialError))

	startTime := time.Now()
	cc, err := grpc.DialContext(ctx, config.Address, dialOptions...)

	if err != nil {
		if config.OnDialError == "panic" {
			logger.Panic("dial grpc server", log.FieldErrKind(errors.ErrKindRequestErr), log.FieldErr(err))
		} else {
			logger.Error("dial grpc server", log.FieldErrKind(errors.ErrKindRequestErr), log.FieldErr(err))
		}
	}
	logger.Info("start grpc client", log.FieldCost(time.Since(startTime)))
	return cc
}

// Build 构建日志
// TODO: 使用log/grpclog
func grpcLogBuild() *log.Logger {
	var (
		once   sync.Once
		logger *log.Logger
	)
	once.Do(func() {
		logger = log.MioLogger.With(log.FieldName("client.grpc"))
	})
	return logger
}
