
# 初始化

config.toml

```toml
[mio.server.grpc]
    port = 9091

[mio.client.directserver]
    address = "127.0.0.1:20102"
    balancerName = "round_robin" # 默认值
    block =  false # 默认值
    dialTimeout = "0s" # 默认值
    debug = true # 开启Debug信息
    disableTrace = false # 开启链路追踪

[mio.trace.jaeger]
    enableRPCMetrics = false
    [mio.trace.jaeger.sampler]
        type = "const"
        param = 0.001
```

mail.go

```go
func main() {
 eng := NewEngine()
 // eng.SetGovernor("127.0.0.1:9092")
 if err := eng.Run(); err != nil {
  xlog.Error(err.Error())
 }
}

type Engine struct {
 mio.Application
}

func NewEngine() *Engine {
 eng := &Engine{}

 if err := eng.Startup(
  eng.serveGRPC,
 ); err != nil {
  xlog.Panic("startup", xlog.Any("err", err))
 }
 return eng
}

func (eng *Engine) serveGRPC() error {
 server := xgrpc.StdConfig("grpc").MustBuild()
 helloworld.RegisterGreeterServer(server.Server, new(Greeter))
 return eng.Serve(server)
}

type Greeter struct{}

func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
 return &helloworld.HelloReply{
  Message: "Hello mio",
 }, nil
}

```
