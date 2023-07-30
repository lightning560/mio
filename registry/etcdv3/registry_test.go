package etcdv3

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"miopkg/client/etcdv3"
	ilog "miopkg/log"
	"miopkg/registry"
	"miopkg/server"
	"miopkg/util/constant"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3/mock/mockserver"
)

func startMockServer() {
	ms, err := mockserver.StartMockServers(1)
	if err != nil {
		log.Fatal(err)
	}

	if err := ms.StartAt(0); err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	go startMockServer()
}

func Test_etcdv3Registry(t *testing.T) {
	etcdConfig := etcdv3.DefaultConfig()
	etcdConfig.Endpoints = []string{"localhost:0"}
	registry, err := newETCDRegistry(&Config{
		Config:      etcdConfig,
		ReadTimeout: time.Second * 10,
		Prefix:      "mio",
		logger:      ilog.DefaultLogger,
	})

	assert.Nil(t, err)
	assert.Nil(t, registry.RegisterService(context.Background(), &server.ServiceInfo{
		Name:       "service_1",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9091",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceProvider,
		Deployment: "default",
		Group:      "",
	}))

	services, err := registry.ListServices(context.Background(), "service_1", "grpc")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, "10.10.10.1:9091", services[0].Address)

	go func() {
		si := &server.ServiceInfo{
			Name:       "service_1",
			Scheme:     "grpc",
			Address:    "10.10.10.1:9092",
			Enable:     true,
			Healthy:    true,
			Metadata:   map[string]string{},
			Region:     "default",
			Zone:       "default",
			Deployment: "default",
		}
		time.Sleep(time.Second)
		assert.Nil(t, registry.RegisterService(context.Background(), si))
		assert.Nil(t, registry.UnregisterService(context.Background(), si))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		endpoints, err := registry.WatchServices(ctx, "service_1", "grpc")
		assert.Nil(t, err)
		for msg := range endpoints {
			t.Logf("watch service: %+v\n", msg)
			// 	assert.Equal(t, "10.10.10.2:9092", msg)
		}
	}()

	time.Sleep(time.Second * 3)
	cancel()
	_ = registry.Close()
	time.Sleep(time.Second * 1)
}

func Test_etcdv3registry_UpdateAddressList(t *testing.T) {
	etcdConfig := etcdv3.DefaultConfig()
	etcdConfig.Endpoints = []string{"localhost:0"}
	reg, err := newETCDRegistry(&Config{
		Config:      etcdConfig,
		ReadTimeout: time.Second * 10,
		Prefix:      "mio",
		logger:      ilog.DefaultLogger,
	})

	assert.Nil(t, err)

	var routeConfig = registry.RouteConfig{
		ID:         "1",
		Scheme:     "grpc",
		Host:       "",
		Deployment: "openapi",
		URI:        "/hello",
		Upstream: registry.Upstream{
			Nodes: map[string]int{
				"10.10.10.1:9091": 1,
				"10.10.10.1:9092": 10,
			},
		},
	}
	_, err = reg.client.Put(context.Background(), "/mio/service_1/configurators/grpc:///routes/1", routeConfig.String())
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		services, err := reg.WatchServices(ctx, "service_1", "grpc")
		assert.Nil(t, err)
		fmt.Printf("len(services) = %+v\n", len(services))
		for service := range services {
			fmt.Printf("service = %+v\n", service)
		}
	}()
	time.Sleep(time.Second * 3)
	cancel()
	_ = reg.Close()
	time.Sleep(time.Second * 1)
}
