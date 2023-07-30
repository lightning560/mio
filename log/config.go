package log

import (
	"fmt"
	"log"
	"time"

	"miopkg/conf"
	"miopkg/env"
	"miopkg/util/constant"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/// 主要定义struct，读取配置

// / 1 init 注册一个conf的Onloaded方法。在conf中执行。RawConfig执行。
// / 2 调用build，生成2个Logger,DefaultLogger & MioLogger
func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init loggers")
		log.Printf("reload default logger with configKey: %s", ConfigEntry("default"))
		///  2 这里读取配置默认有2个key:mio.logger.default mio.logger.mio
		///[mio.logger.default]
		// 		debug = false  # 是否在命令行输出
		// 		enableConsole = false # 是否按命令行格式输出
		// 		name = "default.json"
		// 		dir = "."
		/// 	ConfigPrefix = "mio"
		DefaultLogger = RawConfig(constant.ConfigPrefix + ".logger.default").Build()

		log.Printf("reload default logger with configKey: %s", ConfigEntry("mio"))
		MioLogger = RawConfig(constant.ConfigPrefix + ".logger.Mio").Build()
	})
}

var ConfigPrefix = constant.ConfigPrefix + ".logger"

// Config ...
type Config struct {
	Dir string // Dir 日志输出目录

	Name string // Name 日志文件名称

	Level string // Level 日志初始等级

	Fields []zap.Field // 日志初始化字段

	AddCaller bool // 是否添加调用者信息

	Prefix string // 日志前缀

	MaxSize   int // 日志输出文件最大长度，超过改值则截断
	MaxAge    int
	MaxBackup int

	Interval      time.Duration // 日志磁盘刷盘间隔
	CallerSkip    int
	Async         bool
	Queue         bool
	QueueSleep    time.Duration
	Core          zapcore.Core
	Debug         bool // 是否在命令行输出
	EncoderConfig *zapcore.EncoderConfig
	configKey     string
}

// Filename ...
func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

func ConfigEntry(name string) string {
	return ConfigPrefix + "." + name
}

// RawConfig ...
// / 1 将log配置从conf读取到log中。
// / 如果没有相应的配置就采用默认配置
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	// 调用conf.UnmarshalWithExpect方法将key相应的配置读取到相应的config
	config, _ = conf.UnmarshalWithExpect(key, config).(*Config)
	config.configKey = key
	return config
}

// StdConfig Mio Standard logger config
func StdConfig(name string) *Config {
	return RawConfig(ConfigPrefix + "." + name)
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Name:          "mio_default.json",
		Dir:           env.LogDir(),
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		Interval:      24 * time.Hour,
		CallerSkip:    2,
		AddCaller:     true,
		Async:         true,
		Queue:         false,
		QueueSleep:    100 * time.Millisecond,
		EncoderConfig: DefaultZapConfig(),
		Debug:         true,
		Fields: []zap.Field{
			String("aid", env.AppID()),
			String("iid", env.AppInstance()),
		},
	}
}

// Build ...
// / 2 将配置Config加载到Logger
// / 具体执行的是newLogger
func (config Config) Build() *Logger {
	if config.EncoderConfig == nil {
		config.EncoderConfig = DefaultZapConfig()
	}
	if config.Debug {
		config.EncoderConfig.EncodeLevel = DebugEncodeLevel
	}
	logger := newLogger(&config)
	if config.configKey != "" {
		/// 注册onchanges 一个autolevel方法
		logger.AutoLevel(config.configKey + ".level")
	}
	return logger
}
