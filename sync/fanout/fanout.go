package fanout

import (
	"context"
	"errors"
	"runtime"
	"sync"
	// "miopkg/log"
	// "miopkg/net/metadata"
	// "miopkg/net/trace"
	// "miopkg/stat/prom"
)

// TODO:add metric & trace & metadata
var (
	// ErrFull chan full.
	ErrFull = errors.New("fanout: chan full")
	// stats     = prom.BusinessInfoCount
	// traceTags = []trace.Tag{
	// 	trace.Tag{Key: trace.TagSpanKind, Value: "background"},
	// 	trace.Tag{Key: trace.TagComponent, Value: "sync/pipeline/fanout"},
	// }
)

type options struct {
	worker int
	buffer int
}

// Option fanout option
type Option func(*options)

// Worker specifies the worker of fanout
func Worker(n int) Option {
	if n <= 0 {
		panic("fanout: worker should > 0")
	}
	return func(o *options) {
		o.worker = n
	}
}

// Buffer specifies the buffer of fanout
func Buffer(n int) Option {
	if n <= 0 {
		panic("fanout: buffer should > 0")
	}
	return func(o *options) {
		o.buffer = n
	}
}

type item struct {
	f   func(c context.Context)
	ctx context.Context
}

// Fanout async consume data from chan.
type Fanout struct {
	name    string
	ch      chan item
	options *options
	waiter  sync.WaitGroup

	ctx    context.Context
	cancel func()
}

// 新建一个fanout 对象 名称为cache 名称主要用来上报监控和打日志使用 最好不要重复
// (可选参数) worker数量为1 表示后台只有1个线程在工作
// (可选参数) buffer 为1024 表示缓存chan长度为1024 如果chan满了 再调用Do方法就会报错 设定长度主要为了防止OOM
// New new a fanout struct.
func New(name string, opts ...Option) *Fanout {
	if name == "" {
		name = "fanout"
	}
	o := &options{
		worker: 1,
		buffer: 1024,
	}
	for _, op := range opts {
		op(o)
	}
	c := &Fanout{
		ch:      make(chan item, o.buffer),
		name:    name,
		options: o,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.waiter.Add(o.worker)
	for i := 0; i < o.worker; i++ {
		go c.proc()
	}
	return c
}

// 读channel，异步执行channel传递的函数
func (c *Fanout) proc() {
	defer c.waiter.Done()
	for {
		select {
		case t := <-c.ch:
			wrapFunc(t.f)(t.ctx)
			// stats.State(c.name+"_channel", int64(len(c.ch)))
		case <-c.ctx.Done():
			return
		}
	}
}

// 包装函数，加入recover，防止panic导致程序挂掉
func wrapFunc(f func(c context.Context)) (res func(context.Context)) {
	res = func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64*1024)
				buf = buf[:runtime.Stack(buf, false)]
				// log.Error("panic in fanout proc, err: %s, stack: %s", r, buf)
			}
		}()
		f(ctx)
		// if tr, ok := trace.FromContext(ctx); ok {
		// 	tr.Finish(nil)
		// }
	}
	return
}

// 需要异步执行的方法
// Do save a callback func.
func (c *Fanout) Do(ctx context.Context, f func(ctx context.Context)) (err error) {
	if f == nil || c.ctx.Err() != nil {
		return c.ctx.Err()
	}
	// nakeCtx := metadata.WithContext(ctx)
	// if tr, ok := trace.FromContext(ctx); ok {
	// 	tr = tr.Fork("", "Fanout:Do").SetTag(traceTags...)
	// 	nakeCtx = trace.NewContext(nakeCtx, tr)
	// }
	select {
	case c.ch <- item{f: f, ctx: ctx}:
	default:
		err = ErrFull
	}
	// stats.State(c.name+"_channel", int64(len(c.ch)))
	return
}

// 程序结束的时候关闭fanout 会等待后台线程完成后返回
// Close close fanout
func (c *Fanout) Close() error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.cancel()
	c.waiter.Wait()
	return nil
}
