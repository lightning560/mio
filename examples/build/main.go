package main

import (
	"miopkg/application"
	igin "miopkg/http/gin"
	"miopkg/log"

	"github.com/gin-gonic/gin"
)

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
		eng.serveHTTP,
	); err != nil {
		log.Panic("startup", log.Any("err", err))
	}
	return eng
}

// HTTP地址
func (eng *Engine) serveHTTP() error {
	server := igin.StdConfig("http").Build()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, "Gopher")
		return
	})
	return eng.Serve(server)
}
