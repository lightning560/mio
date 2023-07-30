package main

import (
	"context"

	"miopkg/application"
	"miopkg/examples/grpc/helloworld/helloworld"
	grpcsvr "miopkg/grpc/server"
	"miopkg/log"
	// iglog "miopkg/log/grpclog"
)

// run: go run main.go -config=config.toml
func main() {
	eng := NewEngine()
	// eng.SetGovernor("127.0.0.1:9092")
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
		eng.serveGRPC,
	); err != nil {
		log.Panic("startup", log.Any("err", err))
	}
	return eng
}

func (eng *Engine) serveGRPC() error {
	server := grpcsvr.StdConfig("grpc").MustBuild()
	helloworld.RegisterGreeterServer(server.Server, &Greeter{server: server})
	return eng.Serve(server)
}

type Greeter struct {
	helloworld.UnimplementedGreeterServer
	server *grpcsvr.Server
}

func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: "i'm mio grpc-server ",
	}, nil
}
