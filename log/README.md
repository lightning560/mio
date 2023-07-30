# 使用

## 初始化

在config.go#init中启动
默认初始化2个

1. default
2. mio

## 配置

```go
log.StdConfig("mylog").Build()

// StdConfig Mio Standard logger config
func StdConfig(name string) *Config {
 return RawConfig(ConfigPrefix + "." + name)
}
```

# log

log wrapped go.uber.org/zap, simplify the difficulty of use.

## 动态设置日志级别

修改默认日志级别:

```toml
[mio.logger.default]
    level = "error"
```

修改自定义日志界别:

```toml
[mio.logger.mylog]
    level = "error"
```

## 创建自定义日志

```go
logger := log.StdConfig("mylog").Build()
logger.Info("info", xlog.String("a", "b"))
logger.Infof("info %s", "a")
logger.Infow("info", "a", "b")
```

也可以更精确的控制:

```go
config := log.Config{
    Name: "default.log",
    Dir: "/tmp",
    Level: "info",
}
logger := config.Build()
logger.SetLevel(log.DebugLevel)
logger.Debug("debug", log.String("a", "b"))
logger.Debugf("debug %s", "a")
logger.Debugw("debug", "a", "b")
```
