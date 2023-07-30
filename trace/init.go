package trace

import (
	"log"

	"miopkg/conf"
	"miopkg/trace/jaeger"
)

func init() {
	// 加载完配置，初始化sentinel
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init trace rules")
		if conf.Get("mio.trace.jaeger") != nil {
			var config = jaeger.Load("mio.trace.jaeger")
			SetGlobalTracer(config.Build())
		}
	})
}
