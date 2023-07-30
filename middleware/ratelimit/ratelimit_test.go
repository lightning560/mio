package ratelimit

import (
	"context"
	"errors"
	"fmt"
	xerrors "miopkg/errors"
	"testing"

	"miopkg/middleware/ratelimit/bbr"
)

type (
	ratelimitMock struct {
		reached bool
	}
	ratelimitReachedMock struct {
		reached bool
	}
)

func (r *ratelimitMock) Allow() (bbr.DoneFunc, error) {
	return func(_ bbr.DoneInfo) {
		r.reached = true
	}, nil
}

func (r *ratelimitReachedMock) Allow() (bbr.DoneFunc, error) {
	return func(_ bbr.DoneInfo) {
		r.reached = true
	}, errors.New("errored")
}

func Test_WithLimiter(t *testing.T) {
	o := options{
		limiter: &ratelimitMock{},
	}

	WithLimiter(nil)(&o)
	if o.limiter != nil {
		t.Error("The limiter property must be updated.")
	}
}

func Test_Server(t *testing.T) {
	nextValid := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "Hello valid", nil
	}

	rlm := &ratelimitMock{}
	rlrm := &ratelimitReachedMock{}

	_, _ = Server(func(o *options) {
		o.limiter = rlm
	})(nextValid)(context.Background(), nil)
	if !rlm.reached {
		t.Error("The ratelimit must run the done function.")
	}

	_, _ = Server(func(o *options) {
		o.limiter = rlrm
	})(nextValid)(context.Background(), nil)
	if rlrm.reached {
		t.Error("The ratelimit must not run the done function and should be denied.")
	}
}

func Test_Loop(t *testing.T) {
	allowed := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "Hello valid", nil
	}
	limiter := bbr.NewLimiter()
	for i := 0; i < 1000; i++ {
		done, e := limiter.Allow()
		if e != nil {
			// rejected
			fmt.Println("error", xerrors.ErrLimitExceed)
			// return
		}
		// allowed
		_, err := allowed(context.Background(), nil)
		done(bbr.DoneInfo{Err: err})
	}
}
