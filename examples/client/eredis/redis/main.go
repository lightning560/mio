package main

import (
	"context"
	"errors"
	"fmt"
	"miopkg/application"
	"miopkg/log"

	jsoniter "github.com/json-iterator/go"

	"miopkg/client/eredis"
)

// export MIO_MODE=dev && go run main.go --config=config.toml
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

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func testRedis() error {
	err := eredisClient.SetEX(context.Background(), "hello", "world", 10000)
	fmt.Println("set hello,err:", err)

	v := person{Name: "forgo", Age: 35}
	vjson, _ := jsoniter.MarshalToString(v)
	fmt.Println("vjson:", vjson)
	err = eredisClient.Set(context.Background(), "v", vjson, 0)
	fmt.Println("set v,err:", err)
	str, err := eredisClient.Get(context.Background(), "hello")
	fmt.Println("get hello", str, err)

	str, err = eredisClient.Get(context.Background(), "lee")
	fmt.Println("Get lee", errors.Is(err, eredis.Nil), "err="+err.Error())
	return nil
}
