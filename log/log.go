package log

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"miopkg/conf"

	"miopkg/util/xdefer"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// wrap zap
const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zap.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zap.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-Level logs.
	ErrorLevel = zap.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

// Func ...
type (
	Func  func(string, ...zap.Field)
	Field = zap.Field //包装一下zap
	Level = zapcore.Level
	/// 核心结构体
	Logger struct {
		desugar *zap.Logger
		lv      *zap.AtomicLevel
		config  Config
		sugar   *zap.SugaredLogger
	}
)

// 包装一下zap
var (
	// String ...
	String = zap.String
	// Any ...
	Any = zap.Any
	// Int64 ...
	Int64 = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	// Uint ...
	Uint = zap.Uint
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString
)

/// 3核心，配置zap并且调用zap.core生成
func newLogger(config *Config) *Logger {
	///使用zapOptions，配置zap
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}
	if len(config.Fields) > 0 {
		zapOptions = append(zapOptions, zap.Fields(config.Fields...))
	}
	/// 配置WriteSyncer。debug就stdout,其他模式使用retate写盘
	var ws zapcore.WriteSyncer
	if config.Debug {
		ws = os.Stdout
	} else {
		ws = zapcore.AddSync(newRotate(config))
	}
	/// 是否异步，异步使用buffer,
	if config.Async {
		var close CloseFunc
		ws, close = Buffer(ws, defaultBufferSize, defaultFlushInterval)

		xdefer.Register(close)
	}
	//设置zap的lvl
	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}

	// encoderConfig := defaultZapConfig()
	// if config.Debug {
	// 	encoderConfig = defaultDebugConfig()
	// }
	/// 配置encoder。debug使用console；其他模式使用JSON
	encoderConfig := *config.EncoderConfig
	core := config.Core
	if core == nil {
		core = zapcore.NewCore(
			func() zapcore.Encoder {
				if config.Debug {
					return zapcore.NewConsoleEncoder(encoderConfig)
				}
				return zapcore.NewJSONEncoder(encoderConfig)
			}(),
			ws,
			lv,
		)
	}
	/// 根据之前的core和option，new一个zap,然后放入Logger返回s
	zapLogger := zap.New(
		core,
		zapOptions...,
	)
	return &Logger{
		desugar: zapLogger,
		lv:      &lv,
		config:  *config,
		sugar:   zapLogger.Sugar(),
	}
}

// AutoLevel ...
// / 动态调整zap等级
func (logger *Logger) AutoLevel(confKey string) {
	conf.OnChange(func(config *conf.Configuration) {
		lvText := strings.ToLower(config.GetString(confKey))
		if lvText != "" {
			logger.Info("update level", String("level", lvText), String("name", logger.config.Name))
			_ = logger.lv.UnmarshalText([]byte(lvText))
		}
	})
}

// SetLevel ...
func (logger *Logger) SetLevel(lv Level) {
	logger.lv.SetLevel(lv)
}

// Flush ...
func (logger *Logger) Flush() error {
	return logger.desugar.Sync()
}

// DefaultZapConfig ...
func DefaultZapConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "lv",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// DebugEncodeLevel ...
// / 根据lvl，打印不同的颜色
func DebugEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = color.RedString
	switch lv {
	case zapcore.DebugLevel:
		colorize = color.BlueString
	case zapcore.InfoLevel:
		colorize = color.GreenString
	case zapcore.WarnLevel:
		colorize = color.YellowString
	case zapcore.ErrorLevel, zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
		colorize = color.RedString
	default:
	}
	enc.AppendString(colorize(lv.CapitalString()))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}

// IsDebugMode ...
func (logger *Logger) IsDebugMode() bool {
	return logger.config.Debug
}

func normalizeMessage(msg string) string {
	return fmt.Sprintf("%-32s", msg)
}

// Debug ...
func (logger *Logger) Debug(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Debug(msg, fields...)
}

// Debugw ...
func (logger *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Debugw(msg, keysAndValues...)
}

func sprintf(template string, args ...interface{}) string {
	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	return msg
}

// StdLog ...
func (logger *Logger) StdLog() *log.Logger {
	return zap.NewStdLog(logger.desugar)
}

// Debugf ...
func (logger *Logger) Debugf(template string, args ...interface{}) {
	logger.sugar.Debugw(sprintf(template, args...))
}

// Info ...
func (logger *Logger) Info(msg string, fields ...Field) {
	/// 实现debug的动态配置
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Info(msg, fields...)
}

// Infow ...
func (logger *Logger) Infow(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Infow(msg, keysAndValues...)
}

// Infof ...
func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.sugar.Infof(sprintf(template, args...))
}

// Warn ...
func (logger *Logger) Warn(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Warn(msg, fields...)
}

// Warnw ...
func (logger *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Warnw(msg, keysAndValues...)
}

// Warnf ...
func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.sugar.Warnf(sprintf(template, args...))
}

// Error ...
func (logger *Logger) Error(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Error(msg, fields...)
}

// Errorw ...
func (logger *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Errorw(msg, keysAndValues...)
}

// Errorf ...
func (logger *Logger) Errorf(template string, args ...interface{}) {
	logger.sugar.Errorf(sprintf(template, args...))
}

// Panic ...
func (logger *Logger) Panic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.desugar.Panic(msg, fields...)
}

// Panicw ...
func (logger *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Panicw(msg, keysAndValues...)
}

// Panicf ...
func (logger *Logger) Panicf(template string, args ...interface{}) {
	logger.sugar.Panicf(sprintf(template, args...))
}

// DPanic ...
func (logger *Logger) DPanic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.desugar.DPanic(msg, fields...)
}

// DPanicw ...
func (logger *Logger) DPanicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.DPanicw(msg, keysAndValues...)
}

// DPanicf ...
func (logger *Logger) DPanicf(template string, args ...interface{}) {
	logger.sugar.DPanicf(sprintf(template, args...))
}

// Fatal ...
func (logger *Logger) Fatal(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		_ = normalizeMessage(msg)
		return
	}
	logger.desugar.Fatal(msg, fields...)
}

// Fatalw ...
func (logger *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Fatalw(msg, keysAndValues...)
}

// Fatalf ...
func (logger *Logger) Fatalf(template string, args ...interface{}) {
	logger.sugar.Fatalf(sprintf(template, args...))
}

func panicDetail(msg string, fields ...Field) {
	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}

	// 控制台输出
	fmt.Printf("%s: \n    %s: %s\n", color.RedString("panic"), color.RedString("msg"), msg)
	if _, file, line, ok := runtime.Caller(3); ok {
		fmt.Printf("    %s: %s:%d\n", color.RedString("loc"), file, line)
	}
	for key, val := range enc.Fields {
		fmt.Printf("    %s: %s\n", color.RedString(key), fmt.Sprintf("%+v", val))
	}

}

// With ...
func (logger *Logger) With(fields ...Field) *Logger {
	desugarLogger := logger.desugar.With(fields...)
	return &Logger{
		desugar: desugarLogger,
		lv:      logger.lv,
		sugar:   desugarLogger.Sugar(),
		config:  logger.config,
	}
}

// ZapLogger returns *zap.Logger
func (logger *Logger) ZapLogger() *zap.Logger {
	return logger.desugar
}

// ZapSugaredLogger returns *zap.SugaredLogger
func (logger *Logger) ZapSugaredLogger() *zap.SugaredLogger {
	return logger.sugar
}
