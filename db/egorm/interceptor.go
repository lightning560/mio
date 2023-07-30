package egorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"miopkg/db/egorm/manager"
	ig "miopkg/grpc"

	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/trace"

	ilog "miopkg/log"
	"miopkg/metric"
	itrace "miopkg/trace"
	"miopkg/util/xdebug"

	"gorm.io/gorm"
)

// Handler ...
type Handler func(*gorm.DB)

// Processor ...
type Processor interface {
	Get(name string) func(*gorm.DB)
	Replace(name string, handler func(*gorm.DB)) error
}

// Interceptor ...
type Interceptor func(string, *manager.DSN, string, *config, *ilog.Logger) func(next Handler) Handler

func debugInterceptor(compName string, dsn *manager.DSN, op string, options *config, logger *ilog.Logger) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if !xdebug.IsDevelopmentMode() {
				next(db)
				return
			}
			beg := time.Now()
			next(db)
			cost := time.Since(beg)
			if db.Error != nil {
				log.Println("[egorm.response]",
					ilog.MakeReqResError(0, compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db.Statement.SQL.String(), db.Statement.Vars, true), db.Error.Error()),
				)
			} else {
				log.Println("[egorm.response]",
					ilog.MakeReqResInfo(0, compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db.Statement.SQL.String(), db.Statement.Vars, true), fmt.Sprintf("%v", db.Statement.Dest)),
				)
			}

		}
	}
}

func metricInterceptor(compName string, dsn *manager.DSN, op string, config *config, logger *ilog.Logger) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			beg := time.Now()
			next(db)
			cost := time.Since(beg)

			loggerKeys := ig.CustomContextKeys()

			var fields = make([]ilog.Field, 0, 15+len(loggerKeys))
			fields = append(fields,
				ilog.FieldMethod(op),
				ilog.FieldName(dsn.DBName+"."+db.Statement.Table), ilog.FieldCost(cost))
			if config.EnableAccessInterceptorReq {
				fields = append(fields, ilog.String("req", logSQL(db.Statement.SQL.String(), db.Statement.Vars, config.EnableDetailSQL)))
			}
			if config.EnableAccessInterceptorRes {
				fields = append(fields, ilog.Any("res", db.Statement.Dest))
			}

			// 开启了链路，那么就记录链路id
			if config.EnableTraceInterceptor && itrace.IsGlobalTracerRegistered() {
				fields = append(fields, ilog.FieldTid(itrace.ExtractTraceID(db.Statement.Context)))
			}

			// 支持自定义log
			for _, key := range loggerKeys {
				if value := getContextValue(db.Statement.Context, key); value != "" {
					fields = append(fields, ilog.FieldCustomKeyValue(key, value))
				}
			}

			// 记录监控耗时
			metric.ClientHandleHistogram.WithLabelValues(metric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr).Observe(cost.Seconds())

			// 如果有慢日志，就记录
			if config.SlowLogThreshold > time.Duration(0) && config.SlowLogThreshold < cost {
				logger.Warn("slow", fields...)
			}

			// 如果有错误，记录错误信息
			if db.Error != nil {
				fields = append(fields, ilog.FieldEvent("error"), ilog.FieldErr(db.Error))
				if errors.Is(db.Error, ErrRecordNotFound) {
					logger.Warn("access", fields...)
					metric.ClientHandleCounter.Inc(metric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Empty")
					return
				}
				logger.Error("access", fields...)
				metric.ClientHandleCounter.Inc(metric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Error")
				return
			}

			metric.ClientHandleCounter.Inc(metric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "OK")
			// 开启了记录日志信息，那么就记录access
			// event normal和error，代表全部access的请求数
			if config.EnableAccessInterceptor {
				fields = append(fields,
					ilog.FieldEvent("normal"),
				)
				logger.Info("access", fields...)
			}
		}
	}
}

func logSQL(sql string, args []interface{}, containArgs bool) string {
	if containArgs {
		return bindSQL(sql, args)
	}
	return sql
}

func traceInterceptor(compName string, dsn *manager.DSN, op string, options *config, logger *ilog.Logger) func(Handler) Handler {
	tracer := itrace.NewTracer(trace.SpanKindClient)
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if db.Statement.Context != nil {
				_, span := tracer.Start(db.Statement.Context, "GORM", nil)
				defer span.End()
				// 延迟执行 scope.CombinedConditionSql() 避免sqlVar被重复追加
				next(db)

				span.SetAttributes(
					itrace.String("sql.inner", dsn.DBName),
					itrace.String("sql.addr", dsn.Addr),
					itrace.String("span.kind", "client"),
					itrace.String("peer.service", "mysql"),
					itrace.String("db.instance", dsn.DBName),
					itrace.String("peer.address", dsn.Addr),
					itrace.String("peer.statement", logSQL(db.Statement.SQL.String(), db.Statement.Vars, options.EnableDetailSQL)),
				)
				return
			}

			next(db)
		}
	}
}

func getContextValue(c context.Context, key string) string {
	if key == "" {
		return ""
	}
	return cast.ToString(ig.Value(c, key))
}
