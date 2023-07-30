package metric

import (
	"net/http"
	"time"

	"miopkg/conf"
	"miopkg/env"
	"miopkg/governor"
	"miopkg/util/constant"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics interface {
	prometheus.Registerer
	prometheus.Gatherer

	BulkRegister(...prometheus.Collector) error
}

var (
	// TypeHTTP ...
	TypeHTTP = "http"
	// TypeGRPCUnary ...
	TypeGRPCUnary = "unary"
	// TypeGRPCStream ...
	TypeGRPCStream = "stream"
	// TypeRedis ...
	TypeRedis = "redis"
	TypeGorm  = "gorm"
	// TypeRocketMQ ...
	TypeRocketMQ = "rocketmq"
	// TypeKafka ...
	Typekafka = "kafka"
	// TypeWebsocket ...
	TypeWebsocket = "ws"

	// TypeMySQL ...
	TypeMySQL = "mysql"
	// TypeMySQL ...
	TypeMongo = "mongo"
	// CodeJob
	CodeJobSuccess = "ok"
	// CodeJobFail ...
	CodeJobFail = "fail"
	// CodeJobReentry ...
	CodeJobReentry = "reentry"

	// CodeCache
	CodeCacheMiss = "miss"
	// CodeCacheHit ...
	CodeCacheHit = "hit"
)

var (
	// ServerHandleCounter ...
	ServerHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "server_handle_total",
		Labels:    []string{"type", "method", "client", "code"},
	}.Build()

	// ServerHandleHistogram ...
	ServerHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "server_handle_seconds",
		Labels:    []string{"type", "method", "client"},
	}.Build()

	// ClientHandleCounter ...
	ClientHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "client_handle_total",
		Labels:    []string{"type", "name", "method", "server", "code"},
	}.Build()

	// ClientHandleHistogram ...
	ClientHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "client_handle_seconds",
		Labels:    []string{"type", "name", "method", "server"},
	}.Build()

	// JobHandleCounter ...
	JobHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "job_handle_total",
		Labels:    []string{"type", "name", "code"},
	}.Build()

	// JobHandleHistogram ...
	JobHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "job_handle_seconds",
		Labels:    []string{"type", "name"},
	}.Build()

	LibHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "lib_handle_seconds",
		Labels:    []string{"type", "method", "address"},
	}.Build()
	// LibHandleCounter ...
	LibHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "lib_handle_total",
		Labels:    []string{"type", "method", "address", "code"},
	}.Build()

	LibHandleSummary = SummaryVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "lib_handle_stats",
		Labels:    []string{"name", "status"},
	}.Build()

	// MysqlHandleCounter ...
	MysqlHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "Mysql_handle_total",
		Labels:    []string{"type", "name", "method", "code"},
	}.Build()

	// MysqlHandleHistogram ...
	MysqlHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "Mysql_handle_seconds",
		Labels:    []string{"type", "name", "method"},
	}.Build()

	// MongoHandleCounter ...
	MongoHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "Mongo_handle_total",
		Labels:    []string{"type", "name", "method", "code"},
	}.Build()

	// MongoHandleHistogram ...
	MongoHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "Mongo_handle_seconds",
		Labels:    []string{"type", "name", "method"},
	}.Build()

	// CacheHandleCounter ...
	CacheHandleCounter = CounterVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "cache_handle_total",
		Labels:    []string{"type", "name", "method", "code"},
	}.Build()

	// CacheHandleHistogram ...
	CacheHandleHistogram = HistogramVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "cache_handle_seconds",
		Labels:    []string{"type", "name", "method"},
	}.Build()

	// BuildInfoGauge ...
	BuildInfoGauge = GaugeVecOpts{
		Namespace: constant.DefaultNamespace,
		Name:      "build_info",
		Labels:    []string{"name", "id", "env", "region", "zone", "version", "go_version"},
		// Labels:    []string{"name", "aid", "mode", "region", "zone", "app_version", "mio_version", "start_time", "build_time", "go_version"},
	}.Build()
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		BuildInfoGauge.WithLabelValues(
			env.Name(),
			env.AppID(),
			env.AppMode(),
			env.AppRegion(),
			env.AppZone(),
			env.AppVersion(),
			env.GoVersion(),
		).Set(float64(time.Now().UnixNano() / 1e6))
	})

	governor.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
}
