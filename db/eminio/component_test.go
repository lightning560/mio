package eminio

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"miopkg/conf"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

const (
	bucketName = "my-images"
)

func newCmp() *Component {
	config := `
[minio]
	endpoint = "%s"
	accessKeyID = "%s" 
	secretAccessKey = "%s"
	useSSL = false
`
	config = fmt.Sprintf(config,
		os.Getenv("ENDPOINT"), os.Getenv("ACCESSKEYID"), os.Getenv("SECRETACCESSKEY"),
	)
	if err := conf.LoadFromReader(strings.NewReader(config), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	return Load("minio").Build()
}

func TestComponent_MakeBucketWithContext(t *testing.T) {
	cmp := newCmp()
	ctx := context.Background()
	// location 为空时会采用默认的 region
	err := cmp.MakeBucketWithContext(ctx, bucketName, "")
	assert.Equal(t, nil, err)
}

func TestComponent_ListBucketsWithContext(t *testing.T) {
	cmp := newCmp()
	ctx := context.Background()
	buckets, err := cmp.ListBucketsWithContext(ctx)
	assert.Equal(t, nil, err)
	log.Printf("all buckets:%v \n", buckets)
}

func TestComponent_BucketExistsWithContext(t *testing.T) {
	cmp := newCmp()
	ctx := context.Background()
	exist, err := cmp.BucketExistsWithContext(ctx, bucketName)
	assert.Equal(t, nil, err)
	log.Printf("%s exist:%v \n", bucketName, exist)
}

func TestComponent_RemoveBucket(t *testing.T) {
	cmp := newCmp()
	ctx := context.Background()
	err := cmp.RemoveBucket(ctx, bucketName)
	assert.Equal(t, nil, err)
	exist, err := cmp.BucketExistsWithContext(ctx, bucketName)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, exist)
}
