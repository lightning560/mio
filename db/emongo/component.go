package emongo

import (
	"miopkg/log"
)

const PackageName = "db.mongo"

// Component client (cmdable and config)
type Component struct {
	config *config
	client *Client
	logger *log.Logger
}

// Client returns emongo Client
func (c *Component) Client() *Client {
	return c.client
}
