package eminio

import (
	"fmt"

	"miopkg/conf"
	"miopkg/log"
)

type Option func(c *Container)

type Container struct {
	Config *config // 对外暴露配置
	name   string
	logger *log.Logger
}

// WithRegion 配合region
func WithRegion(region string) Option {
	return func(c *Container) {
		c.Config.Region = region
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := conf.UnmarshalKey(key, &c.Config); err != nil {
		c.logger.Panic("parse config error", log.FieldErr(err), log.FieldKey(key))
		return c
	}
	fmt.Println(c.Config)
	c.logger = c.logger.With(log.FieldMod(key))
	c.name = key
	return c
}

func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.name, c.Config, c.logger)
}

func DefaultContainer() *Container {
	return &Container{
		Config: DefaultConfig(),
		logger: log.MioLogger.With(log.FieldMod(packageName)),
	}
}
