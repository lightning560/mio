package ees

import (
	"miopkg/log"

	"github.com/elastic/go-elasticsearch/v8"
)

const PackageName = "client.es"

// Component ...
type Component struct {
	name   string
	config *config
	logger *log.Logger
	Client *elasticsearch.Client
}

// New ...
func newComponent(name string, config *config, logger *log.Logger) *Component {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     config.Addrs,
		Username:      config.Username,
		Password:      config.Password,
		APIKey:        config.APIKey,
		ServiceToken:  config.ServiceToken,
		RetryOnStatus: config.RetryOnStatus,
		DisableRetry:  !config.EnableRetry,
		// EnableRetryOnTimeout:  config.EnableRetryOnTimeout,//es没有这个属性
		MaxRetries:            config.MaxRetries,
		DiscoverNodesOnStart:  config.EnableDiscoverNodesOnStart,
		DiscoverNodesInterval: config.DiscoverNodesInterval,
		EnableMetrics:         config.EnableMetrics,
		EnableDebugLogger:     config.EnableDebugLogger,
		DisableMetaHeader:     !config.EnableMetaHeader,
	})
	if err != nil {
		logger.Panic("component new panic", log.FieldErr(err))
	}

	cc := &Component{
		name:   name,
		logger: logger,
		config: config,
		Client: client,
	}

	return cc
}
