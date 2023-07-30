package main

import (
	"miopkg/application"
	"miopkg/governor"
	xgin "miopkg/http/gin"
	xlog "miopkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	eng := NewEngine()
	if err := eng.Run(); err != nil {
		xlog.Error(err.Error())
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.serverGovernor,
		eng.serveHTTP,
	); err != nil {
		xlog.Panic("startup", xlog.Any("err", err))
	}
	return eng
}

func (eng *Engine) serverGovernor() error {
	server := governor.StdConfig("governor").Build()
	return eng.Serve(server)
}

// HTTP地址
func (eng *Engine) serveHTTP() error {
	server := xgin.StdConfig("http").Build()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, "Gopher")
	})
	return eng.Serve(server)
}
