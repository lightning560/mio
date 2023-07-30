package log

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/fatih/color"
)

// MakeReqResInfoV2 以info级别打印行号、配置名、目标地址、耗时、请求数据、响应数据
func MakeReqResInfo(callerSkip int, compName string, addr string, cost time.Duration, req interface{}, reply interface{}) string {
	_, file, line, _ := runtime.Caller(callerSkip)
	return fmt.Sprintf("%s %s %s %s %s => %s \n", color.GreenString(file+":"+strconv.Itoa(line)), color.GreenString(compName), color.GreenString(addr), color.YellowString(fmt.Sprintf("[%vms]", float64(cost.Microseconds())/1000)), color.BlueString(fmt.Sprintf("%v", req)), color.BlueString(fmt.Sprintf("%v", reply)))
}

// MakeReqResErrorV2 以error级别打印行号、配置名、目标地址、耗时、请求数据、响应数据
func MakeReqResError(callerSkip int, compName string, addr string, cost time.Duration, req string, err string) string {
	_, file, line, _ := runtime.Caller(callerSkip)
	return fmt.Sprintf("%s %s %s %s %s => %s \n", color.GreenString(file+":"+strconv.Itoa(line)), color.RedString(compName), color.RedString(addr), color.YellowString(fmt.Sprintf("[%vms]", float64(cost.Microseconds())/1000)), color.BlueString(fmt.Sprintf("%v", req)), color.RedString(err))
}
