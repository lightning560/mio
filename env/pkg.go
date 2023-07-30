package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"miopkg/util/xtime"

	"miopkg/util/constant"

	"github.com/fatih/color"
)

const mioVersion = "0.2.0"

var (
	startTime string
	goVersion string
)

// build info
/*

 */
var (
	appName         string
	appID           string
	hostName        string
	buildAppVersion string
	buildUser       string
	buildHost       string
	buildStatus     string
	buildTime       string
)

func init() {
	if appName == "" {
		appName = os.Getenv(constant.EnvAppName)
		if appName == "" {
			appName = filepath.Base(os.Args[0])
		}
	}

	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	hostName = name
	startTime = xtime.TS.Format(time.Now())
	SetBuildTime(buildTime)
	goVersion = runtime.Version()
	InitEnv()
}

// Name gets application name.
func Name() string {
	return appName
}

// SetName set app anme
func SetName(s string) {
	appName = s
}

// AppID get appID
func AppID() string {
	if appID == "" {
		return "1234567890" //default appid when APP_ID Env var not set
	}
	return appID
}

// SetAppID set appID
func SetAppID(s string) {
	appID = s
}

// AppVersion get buildAppVersion
func AppVersion() string {
	return buildAppVersion
}

//appVersion not defined
// func SetAppVersion(s string) {
// 	appVersion = s
// }

// MioVersion get mioVersion
func MioVersion() string {
	return mioVersion
}

// todo: mioVersion is const not be set
// func SetMioVersion(s string) {
// 	mioVersion = s
// }

// BuildTime get buildTime
func BuildTime() string {
	return buildTime
}

// BuildUser get buildUser
func BuildUser() string {
	return buildUser
}

// BuildHost get buildHost
func BuildHost() string {
	return buildHost
}

// SetBuildTime set buildTime
func SetBuildTime(param string) {
	buildTime = strings.Replace(param, "--", " ", 1)
}

// HostName get host name
func HostName() string {
	return hostName
}

// StartTime get start time
func StartTime() string {
	return startTime
}

// GoVersion get go version
func GoVersion() string {
	return goVersion
}

func LogDir() string {
	// LogDir gets application log directory.
	logDir := AppLogDir()
	if logDir == "" {
		if appPodIP != "" && appPodName != "" {
			// k8s 环境
			return fmt.Sprintf("./tmplogs/applogs/%s/%s/", Name(), appPodName)
		}
		return fmt.Sprintf("./tmplogs/applogs/%s/%s/", Name(), appInstance)
	}
	return fmt.Sprintf("%s/%s/%s/", logDir, Name(), appInstance)
}

// PrintVersion print formated version info
func PrintVersion() {
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("name"), color.BlueString(appName))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("appID"), color.BlueString(appID))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("region"), color.BlueString(AppRegion()))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("zone"), color.BlueString(AppZone()))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("appVersion"), color.BlueString(buildAppVersion))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("mioVersion"), color.BlueString(mioVersion))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("buildUser"), color.BlueString(buildUser))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("buildHost"), color.BlueString(buildHost))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("buildTime"), color.BlueString(BuildTime()))
	fmt.Printf("%-8s]> %-30s => %s\n", "mio", color.RedString("buildStatus"), color.BlueString(buildStatus))
}
