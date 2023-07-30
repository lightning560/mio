package xdebug

import (
	"fmt"

	"miopkg/util/xstring"

	"github.com/fatih/color"
	"github.com/tidwall/pretty"
)

// DebugObject ...
func PrintObject(message string, obj interface{}) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%s => %s\n",
		color.RedString(message),
		pretty.Color(
			pretty.Pretty([]byte(xstring.PrettyJson(obj))),
			pretty.TerminalStyle,
		),
	)
}

// DebugBytes ...
func DebugBytes(obj interface{}) string {
	return string(pretty.Color(pretty.Pretty([]byte(xstring.Json(obj))), pretty.TerminalStyle))
}

// PrintKV ...
func PrintKV(key string, val string) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%-50s => %s\n", color.RedString(key), color.GreenString(val))
}

// PrettyKVWithPrefix ...
func PrintKVWithPrefix(prefix string, key string, val string) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%-8s]> %-30s => %s\n", prefix, color.RedString(key), color.BlueString(val))
}

// PrintMap ...
func PrintMap(data map[string]interface{}) {
	if !IsDevelopmentMode() {
		return
	}
	for key, val := range data {
		fmt.Printf("%-20s : %s\n", color.RedString(key), fmt.Sprintf("%+v", val))
	}
}
