package main

import (
	"context"

	"miopkg/application"
	"miopkg/examples/grpc/helloworld/helloworld"
	grpcclt "miopkg/grpc/client"
	"miopkg/log"
	iglog "miopkg/log/grpclog"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// run: go run main.go -config=config.toml
func main() {
	eng := NewEngine()
	if err := eng.Run(); err != nil {
		log.Error(err.Error())
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.consumer,
	); err != nil {
		log.Panic("startup1", log.Any("err", err))
	}
	return eng
}

func (eng *Engine) consumer() error {
	iglog.SetLogger(log.DefaultLogger)
	var headers metadata.MD
	var trailers metadata.MD
	conn := grpcclt.StdConfig("directserver").Build()
	client := helloworld.NewGreeterClient(conn)
	resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "i'm mio grpcclient",
	}, grpc.Header(&headers), grpc.Trailer(&trailers))
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("receive response", log.String("resp", resp.Message))
	}
	spew.Dump(headers)
	spew.Dump(trailers)
	// go func() {
	// 	for {

	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	return nil
}
