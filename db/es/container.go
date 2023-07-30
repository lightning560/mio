package ees

import (
	"miopkg/conf"
	"miopkg/log"
)

type Container struct {
	config *config
	name   string
	logger *log.Logger
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: log.MioLogger.With(log.FieldMod(PackageName)),
	}
}

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

// Build 构建组件
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	cc := newComponent(c.name, c.config, c.logger)
	return cc
}
