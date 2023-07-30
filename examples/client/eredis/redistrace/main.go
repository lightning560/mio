package main

import (
	"context"
	"errors"
	"fmt"
	"miopkg/application"
	"miopkg/log"

	"miopkg/client/eredis"
)

// export Mio_MODE=dev && go run main.go --config=config.toml
type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		invokerRedis,
		testRedis,
	); err != nil {
		log.Panic("Failed to start engine", log.Any("err", err))
	}
	return eng
}

func main() {
	app := NewEngine()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

var eredisClient *eredis.Component

func invokerRedis() error {
	eredisClient = eredis.Load("redis.test").Build()
	return nil
}

func testRedis() error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "X-Mio-Uid", 9527)
	err := eredisClient.Set(ctx, "hello", "world", 0)
	fmt.Println("set hello", err)

	str, err := eredisClient.Get(ctx, "hello")
	fmt.Println("get hello", str, err)

	str, err = eredisClient.Get(ctx, "lee")
	fmt.Println("Get lee", errors.Is(err, eredis.Nil), "err="+err.Error())

	return nil
}
