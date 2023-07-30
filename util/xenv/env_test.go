package xenv

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvOrBoolNoEnv(t *testing.T) {
	flag := EnvOrBool("mio-env-test1", true)
	assert.Equal(t, true, flag)
}

func TestEnvOrBoolHaveEnv(t *testing.T) {
	os.Setenv("mio-env-test1", "false")
	defer os.Unsetenv("mio-env-test1")

	flag := EnvOrBool("mio-env-test1", true)
	assert.Equal(t, false, flag)
}

func TestEnvOrIntNoEnv(t *testing.T) {
	flag := EnvOrInt("mio-env-test1", 1)
	assert.Equal(t, 1, flag)
}

func TestEnvOrIntHaveEnv(t *testing.T) {
	os.Setenv("mio-env-test1", "2")
	defer os.Unsetenv("mio-env-test1")

	flag := EnvOrInt("mio-env-test1", 1)
	assert.Equal(t, 2, flag)
}

func TestEnvOrUintNoEnv(t *testing.T) {
	flag := EnvOrUint("mio-env-test1", 1)
	assert.Equal(t, uint(1), flag)
}

func TestEnvOrUintHaveEnv(t *testing.T) {
	os.Setenv("mio-env-test1", "2")
	defer os.Unsetenv("mio-env-test1")

	flag := EnvOrUint("mio-env-test1", 1)
	assert.Equal(t, uint(2), flag)
}

func TestEnvOrFloat64NoEnv(t *testing.T) {
	flag := EnvOrFloat64("mio-env-test1", 1.1)
	assert.Equal(t, 1.1, flag)
}

func TestEnvOrFloat64HaveEnv(t *testing.T) {
	os.Setenv("mio-env-test1", "1.2")
	defer os.Unsetenv("mio-env-test1")

	flag := EnvOrFloat64("mio-env-test1", 1.1)
	assert.Equal(t, 1.2, flag)
}

func TestEnvOrStrNoEnv(t *testing.T) {
	flag := EnvOrStr("mio-env-test1", "test1")
	assert.Equal(t, "test1", flag)
}

func TestEnvOrStrHaveEnv(t *testing.T) {
	os.Setenv("mio-env-test1", "test2")
	defer os.Unsetenv("mio-env-test1")

	flag := EnvOrStr("mio-env-test1", "test1")
	assert.Equal(t, "test2", flag)
}
