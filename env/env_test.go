package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"miopkg/util/constant"
)

func TestAppMode(t *testing.T) {
	os.Setenv(constant.EnvAppMode, "test-mode")
	defer os.Unsetenv(constant.EnvAppMode)

	InitEnv()
	out := AppMode()
	assert.Equal(t, "test-mode", out)
}
func TestAppRegion(t *testing.T) {
	os.Setenv(constant.EnvAppRegion, "test-region")
	defer os.Unsetenv(constant.EnvAppRegion)

	InitEnv()
	out := AppRegion()
	assert.Equal(t, "test-region", out)
}

func TestAppZone(t *testing.T) {
	os.Setenv(constant.EnvAppZone, "test-zone")
	defer os.Unsetenv(constant.EnvAppZone)

	InitEnv()
	out := AppZone()
	assert.Equal(t, "test-zone", out)
}

func TestAppInstance(t *testing.T) {
	os.Setenv(constant.EnvAppInstance, "test-instance-1")
	defer os.Unsetenv(constant.EnvAppInstance)

	InitEnv()
	out := AppInstance()
	assert.Equal(t, "test-instance-1", out)
}

func TestIsDevelopmentMode(t *testing.T) {
	os.Setenv(constant.MioDebug, "true")
	defer os.Unsetenv(constant.MioDebug)

	InitEnv()
	out := IsDevelopmentMode()
	assert.Equal(t, true, out)
}

func TestMioLogPath(t *testing.T) {
	os.Setenv(constant.MioLogPath, "test-mio.log")
	defer os.Unsetenv(constant.MioLogPath)

	InitEnv()
	out := MioLogPath()
	assert.Equal(t, "test-mio.log", out)
}

func TestEnableLoggerAddApp(t *testing.T) {
	os.Setenv(constant.MioLogAddApp, "true")
	defer os.Unsetenv(constant.MioLogAddApp)

	InitEnv()
	out := EnableLoggerAddApp()
	assert.Equal(t, true, out)
}

func TestMioTraceIDName(t *testing.T) {
	os.Setenv(constant.MioTraceIDName, "x-trace-id")
	defer os.Unsetenv(constant.MioTraceIDName)

	InitEnv()
	out := MioTraceIDName()
	assert.Equal(t, "x-trace-id", out)
}

func TestMioLogExtraKeys(t *testing.T) {
	os.Setenv(constant.MioLogExtraKeys, "x-mio-uid")
	defer os.Unsetenv(constant.MioLogExtraKeys)

	InitEnv()
	out := MioLogExtraKeys()
	assert.Equal(t, []string{"x-mio-uid"}, out)
}

func TestMioLogWriter(t *testing.T) {
	os.Setenv(constant.MioLogWriter, "stderr")
	defer os.Unsetenv(constant.MioLogWriter)
	InitEnv()
	out := MioLogWriter()
	assert.Equal(t, "stderr", out)
}

func TestMioLogEnableAddCaller(t *testing.T) {
	os.Setenv(constant.MioLogEnableAddCaller, "true")
	defer os.Unsetenv(constant.MioLogEnableAddCaller)

	InitEnv()
	out := MioLogEnableAddCaller()
	assert.True(t, out)
}
