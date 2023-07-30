package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomContextKeys(t *testing.T) {
	Set([]string{"X-MIO-Test"})
	arr := CustomContextKeys()
	assert.Equal(t, []string{"X-MIO-Test"}, arr)
	length := CustomContextKeysLength()
	assert.Equal(t, 1, length)
}

func TestValue(t *testing.T) {
	Set([]string{"X-MIO-Test"})
	ctx := context.Background()
	ctx = WithValue(ctx, "X-MIO-Test", "hello")
	val := Value(ctx, "X-MIO-Test")
	assert.Equal(t, "hello", val)
}

func Test_newContextKey(t *testing.T) {
	key := newContextKey("hello")
	assert.Equal(t, "mio context value hello", key.String())
}
