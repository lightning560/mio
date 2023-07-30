package main

import (
	"context"
	"fmt"
	"time"

	"miopkg/application"
	"miopkg/client/etcdv3"
	xlog "miopkg/log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	eng := &application.Application{}
	err := eng.Startup(
		func() error {
			client := etcdv3.StdConfig("myetcd").MustBuild()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			// 添加数据
			_, err := client.Put(ctx, "/hello", "mio")
			xlog.Debug("etcd Put /hello:mio")
			if err != nil {
				xlog.Panic(err.Error())
			}

			// 获取数据
			response, err := client.Get(ctx, "/hello", clientv3.WithPrefix())
			fmt.Println("etcd Get /hello:%s", string(response.Kvs[0].Value))
			if err != nil {
				xlog.Panic(err.Error())
			}
			xlog.Info("get etcd info", xlog.String("key", string(response.Kvs[0].Key)), xlog.String("value", string(response.Kvs[0].Value)))
			return nil
		},
	)
	if err != nil {
		xlog.Panic("startup", xlog.Any("err", err))
	}
}
