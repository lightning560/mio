package p2c

import (
	"miopkg/util/xp2c"
	"miopkg/util/xp2c/leastloaded"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	_ "google.golang.org/grpc/health"
)

// Name is the name of p2c with least loaded balancer.
const (
	Name = "p2c_least_loaded"
)

// newBuilder creates a new balance builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &p2cPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type p2cPickerBuilder struct{}

func (*p2cPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	grpclog.Infof("p2cPickerBuilder: newPicker called with readySCs: %v", info.ReadySCs)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var p2c = leastloaded.New()

	for sc := range info.ReadySCs {
		p2c.Add(sc)
	}

	rp := &p2cPicker{
		p2c: p2c,
	}
	return rp
}

type p2cPicker struct {
	p2c xp2c.P2c
}

// Pick ...
func (p *p2cPicker) Pick(opts balancer.PickInfo) (balancer.PickResult, error) {
	item, done := p.p2c.Next()
	if item == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	return balancer.PickResult{SubConn: item.(balancer.SubConn), Done: done}, nil
}
