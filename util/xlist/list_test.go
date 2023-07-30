package xlist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64Difference(t *testing.T) {
	all := []int64{1, 2, 3, 4, 5}
	minus := []int64{1, 2}
	diff := Int64Difference(all, minus)
	t.Log(diff)
	fmt.Println(diff)
	assert.Equal(t, []int64{3, 4, 5}, diff)
}
