package egorm

import (
	"fmt"

	"miopkg/conf"
	_ "miopkg/db/egorm/internal/dsn"
	"miopkg/db/egorm/manager"
	"miopkg/log"
	"miopkg/metric"
)

// Container ...
type Container struct {
	config    *config
	name      string
	logger    *log.Logger
	dsnParser manager.DSNParser
}

// DefaultContainer ...
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: log.MioLogger.With(log.FieldMod(PackageName)),
	}
}

// Load ...
func Load(key string) *Container {
	c := DefaultContainer()
	if err := conf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", log.FieldErr(err), log.FieldKey(key))
		return c
	}

	c.logger = c.logger.With(log.FieldMod(key))
	c.name = key
	return c
}

func (c *Container) setDSNParser(dialect string) error {
	dsnParser := manager.Get(dialect)
	if dsnParser == nil {
		return fmt.Errorf("invalid support Dialect: %s", dialect)
	}
	c.dsnParser = dsnParser
	return nil
}

// Build 构建组件
func (c *Container) Build(options ...Option) *Component {
	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor))
	}

	if c.config.EnableTraceInterceptor {
		options = append(options, WithInterceptor(traceInterceptor))
	}

	if c.config.EnableMetricInterceptor {
		options = append(options, WithInterceptor(metricInterceptor))
	}

	for _, option := range options {
		option(c)
	}

	var err error
	// todo 设置补齐超时时间
	// timeout 1s
	// readTimeout 5s
	// writeTimeout 5s
	err = c.setDSNParser(c.config.Dialect)
	if err != nil {
		c.logger.Panic("setDSNParser err", log.String("dialect", c.config.Dialect), log.FieldErr(err))
	}
	c.config.dsnCfg, err = c.dsnParser.ParseDSN(c.config.DSN)

	if err == nil {
		c.logger.Info("start db", log.FieldAddr(c.config.dsnCfg.Addr), log.FieldName(c.config.dsnCfg.DBName))
	} else {
		c.logger.Panic("start db", log.FieldErr(err))
	}

	c.logger = c.logger.With(log.FieldAddr(c.config.dsnCfg.Addr))

	component, err := newComponent(c.name, c.dsnParser, c.config, c.logger)
	if err != nil {
		if c.config.OnFail == "panic" {
			c.logger.Panic("open db", log.FieldErrKind("register err"), log.FieldErr(err), log.FieldAddr(c.config.dsnCfg.Addr), log.FieldValueAny(c.config))
		} else {
			metric.ClientHandleCounter.Inc(metric.TypeGorm, c.name, c.name+".ping", c.config.dsnCfg.Addr, "open err")
			c.logger.Error("open db", log.FieldErrKind("register err"), log.FieldErr(err), log.FieldAddr(c.config.dsnCfg.Addr), log.FieldValueAny(c.config))
			return component
		}
	}

	sqlDB, err := component.DB()
	if err != nil {
		c.logger.Panic("ping db", log.FieldErrKind("register err"), log.FieldErr(err), log.FieldValueAny(c.config))
	}
	if err := sqlDB.Ping(); err != nil {
		c.logger.Panic("ping db", log.FieldErrKind("register err"), log.FieldErr(err), log.FieldValueAny(c.config))
	}

	// store db
	instances.Store(c.name, component)
	return component
}
