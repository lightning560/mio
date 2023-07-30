package eredis

import (
	"context"
	"strings"
	"testing"

	"miopkg/conf"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func newCmp() *Component {
	config := `
[redis]
	mode = "sentinel"
	masterName = "redis-master"
	addrs = ["localhost:26379","localhost:26380","localhost:26380"]
`
	if err := conf.LoadFromReader(strings.NewReader(config), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	cmp := Load("redis").Build()
	return cmp
}

func TestSentinel(t *testing.T) {
	cmp := newCmp()
	res, err := cmp.Ping(context.TODO())
	assert.NoError(t, err)
	t.Log("ping result", res)
}
