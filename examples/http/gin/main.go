package main

import (
	"log"

	"miopkg/application"
	xgin "miopkg/http/gin"
	xlog "miopkg/log"

	"github.com/gin-gonic/gin"
)

// go run main.go --config=config.toml
func main() {
	eng := NewEngine()
	if err := eng.Run(); err != nil {
		xlog.Panic(err.Error())
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
		xlog.Panic("startup", xlog.Any("err", err))
	}
	return eng
}

// HTTP地址
func (eng *Engine) serveHTTP() error {
	server := xgin.StdConfig("http").Build()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200, "Hello Gin")
	})
	//Upgrade to websocket
	server.Upgrade(xgin.WebSocketOptions("/ws", func(ws xgin.WebSocketConn, err error) {
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = ws.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))
	return eng.Serve(server)
}
