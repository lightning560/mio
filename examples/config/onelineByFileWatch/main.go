package main

import (
	"time"

	"miopkg/application"
	"miopkg/conf"
	_ "miopkg/conf/datasource/file"
	"miopkg/http/gin"
	"miopkg/log"
)

// go run main.go --config=config.toml --watch=true
func main() {
	app := NewEngine()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.fileWatch,
		eng.serveHTTP,
	); err != nil {
		log.Panic("startup", log.Any("err", err))
	}

	return eng
}

func (eng *Engine) serveHTTP() error {
	server := gin.StdConfig("http").Build()
	return eng.Serve(server)
}

func (s *Engine) fileWatch() error {
	log.DefaultLogger = log.StdConfig("default").Build()
	go func() {
		// 循环打印配置
		for {
			time.Sleep(10 * time.Second)
			peopleName := conf.GetString("people.name")
			log.Info("people info", log.String("name", peopleName), log.String("type", "onelineByFileWatch"))
		}
	}()
	return nil
}
