package constant

const (
	// EnvKeySentinelLogDir ...
	EnvKeySentinelLogDir = "SENTINEL_LOG_DIR"
	// EnvKeySentinelAppName ...
	EnvKeySentinelAppName = "SENTINEL_APP_NAME"
)

const (
	// EnvAppName defines your application name.
	EnvAppName = "APP_NAME"
	EnvAppID   = "APP_ID"

	EnvDeployment = "APP_DEPLOYMENT"

	EnvAppLogDir = "APP_LOG_DIR"
	// EnvAppMode defines your application running mode, you can set APP_MODE with words such as "development/testing/production".
	EnvAppMode = "APP_MODE"
	// EnvAppRegion defines your application running region, such as "ASIA".
	EnvAppRegion = "APP_REGION"
	// EnvAppZone defines your application running zone "ShangHai2".
	EnvAppZone = "APP_ZONE"
	// EnvAppHost defines your application running HOST name.
	EnvAppHost = "APP_HOST"
	// EnvAppInstance defines your application replication unique instance ID.
	EnvAppInstance = "APP_INSTANCE"
	//k8s环境
	EnvPOD_IP   = "POD_IP"
	EnvPOD_NAME = "POD_NAME"
)
const (
	// MioDebug when set to true, Mio will print verbose logs.
	MioDebug = "MIO_DEBUG"
	// MioConfigPath defines your application configuration dsn.
	MioConfigPath = "MIO_CONFIG_PATH"
	// MioLogPath if your application used file writer logger to print logs, MioLogPath is logs directory.
	MioLogPath = "MIO_LOG_PATH"
	// MioLogAddApp defines if we should append application name to every logger entries.
	MioLogAddApp = "MIO_LOG_ADD_APP"
	// MioLogExtraKeys used to append extra tracing keys to every access logger entries, the keys usually comes from HTTP Headers or gRPC Metadata.
	// you can trace you custom business clues, such as "X-Biz-Uid"(your application user ID) or "X-Biz-Order-Id"(your application order ID).
	// each keys separated with ",". For example, export MIO_LOG_EXTRA_KEYS=X-Mio-Uid,X-Mio-Order-Id
	MioLogExtraKeys = "MIO_LOG_EXTRA_KEYS"
	// MioLogWriter defines your log writer, available types are: "file/stderr"
	MioLogWriter = "MIO_LOG_WRITER"
	// MioLogTimeType defines time format on your logger entries, available types are "second/millisecond/%Y-%m-%d %H:%M:%S"
	MioLogTimeType = "MIO_LOG_TIME_TYPE"
	// MioTraceIDName defines tracing ID NAME, default value is "x-trace-id"
	MioTraceIDName = "MIO_TRACE_ID_NAME"
	// MioGovernorEnableConfig defines if you can query current configuration with governor APIs.
	MioGovernorEnableConfig = "MIO_GOVERNOR_ENABLE_CONFIG"
	// MioLogEnableAddCaller when set to true, your log will show caller, default value is false
	MioLogEnableAddCaller = "MIO_LOG_ENABLE_ADD_CALLER"
)

const (
	// DefaultDeployment ...
	DefaultDeployment = ""
	// DefaultRegion ...
	DefaultRegion = ""
	// DefaultZone ...
	DefaultZone = ""
)

const (
	// KeyBalanceGroup ...
	KeyBalanceGroup = "__group"

	// DefaultBalanceGroup ...
	DefaultBalanceGroup = "default"
)
