package env

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"

	"miopkg/util/constant"
	"miopkg/util/xenv"
)

var (
	appLogDir               string
	appMode                 string
	appRegion               string
	appZone                 string
	appHost                 string
	appInstance             string // 通常是实例的机器名
	appPodIP                string
	appPodName              string
	mioDebug                string
	mioLogPath              string
	mioLogAddApp            string
	mioTraceIDName          string
	mioLogExtraKeys         []string
	mioLogWriter            string
	mioGovernorEnableConfig string
	mioLogTimeType          string
	mioLogEnableAddCaller   bool
)

func InitEnv() {
	appID = os.Getenv(constant.EnvAppID)
	appLogDir = os.Getenv(constant.EnvAppLogDir)
	appMode = os.Getenv(constant.EnvAppMode)
	appRegion = os.Getenv(constant.EnvAppRegion)
	appZone = os.Getenv(constant.EnvAppZone)
	appHost = os.Getenv(constant.EnvAppHost)
	appInstance = os.Getenv(constant.EnvAppInstance)
	if appInstance == "" {
		appInstance = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", HostName(), AppID()))))
	}
	appPodIP = os.Getenv(constant.EnvPOD_IP)
	appPodName = os.Getenv(constant.EnvPOD_NAME)

	mioDebug = os.Getenv(constant.MioDebug)
	mioLogPath = os.Getenv(constant.MioLogPath)
	mioLogAddApp = os.Getenv(constant.MioLogAddApp)
	mioTraceIDName = xenv.EnvOrStr(constant.MioTraceIDName, "x-trace-id")
	mioGovernorEnableConfig = os.Getenv(constant.MioGovernorEnableConfig)
	if envMioLogExtraKeys := strings.TrimSpace(os.Getenv(constant.MioLogExtraKeys)); envMioLogExtraKeys != "" {
		mioLogExtraKeys = strings.Split(envMioLogExtraKeys, ",")
	}
	mioLogWriter = xenv.EnvOrStr(constant.MioLogWriter, "file")
	mioLogTimeType = xenv.EnvOrStr(constant.MioLogTimeType, "second")
	if IsDevelopmentMode() {
		mioLogTimeType = "%Y-%m-%d %H:%M:%S"
	}
	mioLogEnableAddCaller = xenv.EnvOrBool(constant.MioLogEnableAddCaller, false)
}

func AppLogDir() string {
	return appLogDir
}

func SetAppLogDir(logDir string) {
	appLogDir = logDir
}

func AppMode() string {
	return appMode
}

func SetAppMode(mode string) {
	appMode = mode
}

func AppRegion() string {
	return appRegion
}

func SetAppRegion(region string) {
	appRegion = region
}

func AppZone() string {
	return appZone
}

func SetAppZone(zone string) {
	appZone = zone
}

func AppHost() string {
	return appHost
}

func SetAppHost(host string) {
	appHost = host
}

func AppInstance() string {
	return appInstance
}

func SetAppInstance(instance string) {
	appInstance = instance
}

// IsDevelopmentMode returns flag if application is in debug mode.
func IsDevelopmentMode() bool {
	return mioDebug == "true"
}

// mioLogPath returns application log file directory path when user choose to write log fo file.
func MioLogPath() string {
	return mioLogPath
}

// EnableLoggerAddApp returns flag if logger has append app Field to log entry.
func EnableLoggerAddApp() bool {
	return mioLogAddApp == "true"
}

// MioTraceIDName returns the key in Metadata for storing traceID
func MioTraceIDName() string {
	return mioTraceIDName
}

// MioLogExtraKeys returns custom extra log keys.
func MioLogExtraKeys() []string {
	return mioLogExtraKeys
}

// MioLogWriter ...
func MioLogWriter() string {
	return mioLogWriter
}

// MioGovernorEnableConfig ...
func MioGovernorEnableConfig() bool {
	return mioGovernorEnableConfig == "true"
}

// MioLogTimeType ...
func MioLogTimeType() string {
	return mioLogTimeType
}

// SetMioDebug returns the flag if debug mode has been triggered
func SetmioDebug(flag string) {
	mioDebug = flag
}

// MioLogEnableAddCaller ...
func MioLogEnableAddCaller() bool {
	return mioLogEnableAddCaller
}
