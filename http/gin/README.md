# 使用

没有init

## 配置

RawConfig("mio.server." + name)
ip使用的`flag.String("host")`

## 初始化

example/http/gin/main.go

```go
import (
 "log"
 "github.com/gin-gonic/gin"
)

func main() {
 eng := NewEngine()
 if err := eng.Run(); err != nil {
  xlog.Panic(err.Error())
 }
}

type Engine struct {
 mio.Application
}

func NewEngine() *Engine {
 eng := &Engine{}
 if err := eng.Startup(eng.serveHTTP,); err != nil {
  xlog.Panic("startup", xlog.Any("err", err))
 }
 return eng
}

// HTTP地址
/// 这里启gin服务
/// 然后将启动的服务灌入eng.Startup( eng.serveHTTP,)
func (eng *Engine) serveHTTP() error {
 server := xgin.StdConfig("http").Build()
 server.GET("/hello", func(ctx *gin.Context) {
  ctx.JSON(200, "Hello Gin")
 })
 //Upgrade to websocket
 // 这里启动ws
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

```
