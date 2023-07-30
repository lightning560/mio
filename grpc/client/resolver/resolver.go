package resolver

import (
	"context"

	"miopkg/registry"
	"miopkg/util/constant"
	"miopkg/util/xgo"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// Register ...
func Register(name string, reg registry.Registry) {
	resolver.Register(&baseBuilder{
		name: name,
		reg:  reg,
	})
}

type baseBuilder struct {
	name string
	reg  registry.Registry
}

// grpc新版本1.40以后，会采用url parse解析，获取endpoint，但这个方法官方说了会有些问题。
// 而 Mio 支持 unix socket，所以需要做一些兼容处理，详情请看 grpc.ClientConn.parseTarget 方法
// For targets of the form "[scheme]://[authority]/endpoint, the endpoint
// value returned from url.Parse() contains a leading "/". Although this is
// in accordance with RFC 3986, we do not want to break existing resolver
// implementations which expect the endpoint without the leading "/". So, we
// end up stripping the leading "/" here. But this will result in an
// incorrect parsing for something like "unix:///path/to/socket". Since we
// own the "unix" resolver, we can workaround in the unix resolver by using
// the `URL` field instead of the `Endpoint` field.

// Build ...
func (b *baseBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	endpoints, err := b.reg.WatchServices(context.Background(), target.Endpoint, "grpc")
	if err != nil {
		return nil, err
	}

	var stop = make(chan struct{})
	xgo.Go(func() {
		for {
			select {
			case endpoint := <-endpoints:
				var state = resolver.State{
					Addresses: make([]resolver.Address, 0),
					Attributes: attributes.
						New(constant.KeyRouteConfig, endpoint.RouteConfigs).             // 路由配置
						WithValue(constant.KeyProviderConfig, endpoint.ProviderConfigs). // 服务提供方元信息
						WithValue(constant.KeyConsumerConfig, endpoint.ConsumerConfigs), // 服务消费方配置信息,
				}
				for _, node := range endpoint.Nodes {
					var address resolver.Address
					address.Addr = node.Address
					address.ServerName = target.Endpoint
					address.Attributes = attributes.New(constant.KeyServiceInfo, node)
					state.Addresses = append(state.Addresses, address)
				}
				_ = cc.UpdateState(state)
			case <-stop:
				return
			}
		}
	})

	return &baseResolver{
		stop: stop,
	}, nil
}

// Scheme ...
func (b baseBuilder) Scheme() string {
	return b.name
}

type baseResolver struct {
	stop chan struct{}
}

// ResolveNow ...
func (b *baseResolver) ResolveNow(options resolver.ResolveNowOptions) {}

// Close ...
func (b *baseResolver) Close() { b.stop <- struct{}{} }
