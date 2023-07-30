package emongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	ilog "miopkg/log"
	"miopkg/metric"
	"miopkg/util/xdebug"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	metricType = "mongo"
)

type Interceptor func(oldProcessFn processFn) (newProcessFn processFn)

func InterceptorChain(interceptors ...Interceptor) func(oldProcess processFn) processFn {
	build := func(interceptor Interceptor, oldProcess processFn) processFn {
		return interceptor(oldProcess)
	}

	return func(oldProcess processFn) processFn {
		chain := oldProcess
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain)
		}
		return chain
	}
}

func debugInterceptor(compName string, c *config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			if !xdebug.IsDevelopmentMode() {
				return oldProcess(cmd)
			}
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if err != nil {
				log.Println("[emongo.response]", ilog.MakeReqResError(0, compName,
					fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), err.Error()),
				)
			} else {
				log.Println("[emongo.response]", ilog.MakeReqResInfo(0, compName,
					fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), fmt.Sprintf("%v", cmd.res)),
				)
			}
			return err
		}
	}
}

func metricInterceptor(compName string, c *config, logger *ilog.Logger) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					metric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "Empty")
				} else {
					metric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "Error")
				}
			} else {
				metric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "OK")
			}
			metric.ClientHandleHistogram.WithLabelValues(metricType, compName, cmd.name, c.DSN).Observe(cost.Seconds())
			return err
		}
	}
}

func accessInterceptor(compName string, c *config, logger *ilog.Logger) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)

			var fields = make([]ilog.Field, 0, 15)
			fields = append(fields,
				ilog.FieldMod(compName),
				ilog.FieldMethod(cmd.name),
				ilog.FieldCost(cost),
			)
			if c.EnableAccessInterceptorReq {
				fields = append(fields, ilog.Any("req", cmd.req))
			}
			if c.EnableAccessInterceptorRes && err == nil {
				fields = append(fields, ilog.Any("res", cmd.res))
			}

			if c.SlowLogThreshold > time.Duration(0) && cost > c.SlowLogThreshold {
				logger.Warn("slow", fields...)
			}

			if err != nil {
				fields = append(fields, ilog.FieldEvent("error"), ilog.FieldErr(err))
				if errors.Is(err, mongo.ErrNoDocuments) {
					logger.Warn("access", fields...)
					return err
				}
				logger.Error("access", fields...)
				return err
			}

			if c.EnableAccessInterceptor {
				fields = append(fields, ilog.FieldEvent("normal"))
				logger.Info("access", fields...)
			}
			return nil
		}
	}
}

func emptyFilterInterceptor(compName string, c *config, logger *ilog.Logger) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			// fmt.Println("filterInterceptor cmd.name:", cmd.name)
			var fields = make([]ilog.Field, 0, 15)
			err := oldProcess(cmd)
			fmt.Println("filterInterceptor cmd.name:", cmd.name)
			fmt.Println("filterInterceptor cmd.req:", cmd.req)
			fmt.Println("filterInterceptor cmd.res:", cmd.res)
			if err != nil {
				fields = append(fields, ilog.FieldEvent("error"), ilog.FieldErr(err))
				if errors.Is(err, mongo.ErrNoDocuments) {
					logger.Warn("access", fields...)
					return err
				}
				logger.Error("access", fields...)
				return err
			}
			return nil
		}
	}
}

func circuitBreakerInterceptor(compName string, c *config, logger *ilog.Logger) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			var fields = make([]ilog.Field, 0, 15)
			err := oldProcess(cmd)
			fmt.Println("filterInterceptor cmd.name:", cmd.name)
			fmt.Println("filterInterceptor cmd.req:", cmd.req)
			fmt.Println("filterInterceptor cmd.res:", cmd.res)
			if err != nil {
				fields = append(fields, ilog.FieldEvent("error"), ilog.FieldErr(err))
				if errors.Is(err, mongo.ErrNoDocuments) {
					logger.Warn("access", fields...)
					return err
				}
				logger.Error("access", fields...)
				return err
			}
			return nil
		}
	}
}

func mustJsonMarshal(val interface{}) string {
	res, _ := json.Marshal(val)
	return string(res)
}
