package eredbloom

import (
	redisbloom "github.com/RedisBloom/redisbloom-go"
)

type Redbloom struct {
	*redisbloom.Client
	config *config
}

func NewRedbloomClient(config *config) *redisbloom.Client {
	return redisbloom.NewClient(config.Address, config.Name, config.AuthPass)
}

func BuildRedbloom() *Redbloom {
	config := LoadConfig("mio.redbloom")
	client := NewRedbloomClient(config)
	return &Redbloom{
		Client: client,
		config: config,
	}
}

