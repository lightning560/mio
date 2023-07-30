
# init

flag 注册一个"version",仅仅用于env.PrintVersion()

# 使用

```go
func main() {
 eng := NewEngine()
 // eng.SetGovernor("127.0.0.1:9092")
 /// 最后一步。为整个应用注入生命周期
 if err := eng.Run(); err != nil {
  xlog.Error(err.Error())
 }
}
//1 包装一下application
type Engine struct {
 mio.Application
}

func NewEngine() *Engine {
    // 2 实例化一个Engine
 eng := &Engine{}
/// 6 启动eng\app,主要是装入服务svr
 if err := eng.Startup(
  eng.serveGRPC,
 ); err != nil {
  xlog.Panic("startup", xlog.Any("err", err))
 }
 return eng
}

/// 3 创建服务
func (eng *Engine) serveGRPC() error {
 server := xgrpc.StdConfig("grpc").MustBuild()
 // 4 注册grpcsvr连接和 Greeter已经实现了
 /// helloworld就是helloworld.pb.go
 helloworld.RegisterGreeterServer(server.Server, new(Greeter))
 /// 5 server等于start
 return eng.Serve(server)
}

type Greeter struct{}

func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
 return &helloworld.HelloReply{
  Message: "Hello mio",
 }, nil
}

```
