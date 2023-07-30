package ekafka

import (
	"fmt"

	"miopkg/conf"
	"miopkg/log"
)

type Option func(c *Container)

type Container struct {
	config *config
	name   string
	logger *log.Logger
}

// DefaultContainer 返回默认Container
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: log.MioLogger.With(log.FieldMod(PackageName)),
	}
}

// Load 载入配置，初始化Container
func Load(key string) *Container {
	c := DefaultContainer()
	// if err := conf.UnmarshalKey(key, &c.config, conf.WithWeaklyTypedInput(true)); err != nil {
	if err := conf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", log.FieldErr(err), log.FieldKey(key))
		return c
	}

	c.logger = c.logger.With(log.FieldMod(key))
	c.name = key
	return c
}

// Build 构建Container
func (c *Container) Build(options ...Option) *Component {
	// 放第一个时间才准确
	options = append(options, WithClientInterceptor(fixedClientInterceptor(c.name, c.config)))
	options = append(options, WithClientInterceptor(traceClientInterceptor(c.name, c.config)))
	options = append(options, WithClientInterceptor(accessClientInterceptor(c.name, c.config, c.logger)))
	if c.config.EnableMetricInterceptor {
		options = append(options, WithClientInterceptor(metricClientInterceptor(c.name, c.config)))
	}

	options = append(options, WithServerInterceptor(fixedServerInterceptor(c.name, c.config)))
	options = append(options, WithServerInterceptor(traceServerInterceptor(c.name, c.config)))
	options = append(options, WithServerInterceptor(accessServerInterceptor(c.name, c.config, c.logger)))
	if c.config.EnableMetricInterceptor {
		options = append(options, WithServerInterceptor(metricServerInterceptor(c.name, c.config)))
	}

	for _, option := range options {
		option(c)
	}

	c.logger = c.logger.With(log.FieldAddr(fmt.Sprintf("%s", c.config.Brokers)))
	cmp := &Component{
		config:         c.config,
		logger:         c.logger,
		producers:      make(map[string]*Producer),
		consumers:      make(map[string]*Consumer),
		consumerGroups: make(map[string]*ConsumerGroup),
		compName:       c.name,
	}

	return cmp
}
