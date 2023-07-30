package consumerserver

import (
	"miopkg/conf"
	"miopkg/log"
)

type Option func(c *Container)

type Container struct {
	name   string
	config *config
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
	for _, option := range options {
		option(c)
	}

	cmp := NewConsumerServerComponent(
		c.name,
		c.config,
		c.config.ekafkaComponent,
		c.logger,
	)

	return cmp
}
