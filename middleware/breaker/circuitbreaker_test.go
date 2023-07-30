package circuitbreaker

import (
	"context"
	"fmt"
	"testing"

	"miopkg/errors"
	"miopkg/middleware/breaker/sre"
	"miopkg/util/group"
	"miopkg/util/transport"
)

type transportMock struct {
	kind      string
	endpoint  string
	operation string
}

type circuitBreakerMock struct {
	err error
}

func (tr *transportMock) Kind() string {
	return tr.kind
}

func (tr *transportMock) Endpoint() string {
	return tr.endpoint
}

func (tr *transportMock) Operation() string {
	return tr.operation
}

func (tr *transportMock) RequestHeader() transport.Header {
	return nil
}

func (tr *transportMock) ReplyHeader() transport.Header {
	return nil
}

func (c *circuitBreakerMock) Allow() error { return c.err }
func (c *circuitBreakerMock) MarkSuccess() {}
func (c *circuitBreakerMock) MarkFailed()  {}

func Test_WithGroup(t *testing.T) {
	o := options{
		group: group.NewGroup(func() interface{} {
			return ""
		}),
	}

	WithGroup(nil)(&o)
	if o.group != nil {
		t.Error("The group property must be updated to nil.")
	}
}

func Test_AllowedServerWithoutGroup(t *testing.T) {
	// allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
	// 	return "Hello valid", nil
	// }
	faildFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.InternalServer("", "")
	}
	breaker := sre.NewBreaker()
	if err := breaker.Allow(); err != nil {
		// rejected
		// NOTE: when client reject requets locally,
		// continue add counter let the drop ratio higher.
		fmt.Println("Error: ", err)
		breaker.MarkFailed()
		return
	}
	_, err := faildFunc(context.Background(), nil)
	if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
		breaker.MarkFailed()
	} else {
		breaker.MarkSuccess()
	}
}

func Test_LoopAllowed(t *testing.T) {
	allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "Hello valid", nil
	}
	fmt.Println("Start loop")
	breaker := sre.NewBreaker()
	for i := 0; i < 1000; i++ {
		if err := breaker.Allow(); err != nil {
			// rejected
			// NOTE: when client reject requets locally,
			// continue add counter let the drop ratio higher.
			fmt.Println("Error: ", err)
			breaker.MarkFailed()
			return
		}
		// allowed
		_, err := allowedFunc(context.Background(), nil)
		if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
	}
}

func Test_LoopFailed(t *testing.T) {
	// allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
	// 	return "Hello valid", nil
	// }
	faildFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.InternalServer("", "")
	}
	fmt.Println("Start loop")
	breaker := sre.NewBreaker()
	count := 0
	for i := 0; i < 1000; i++ {
		if err := breaker.Allow(); err != nil {
			// rejected
			// NOTE: when client reject requets locally,
			// continue add counter let the drop ratio higher.
			fmt.Println("Error: ", err)
			count++
			fmt.Println("count: ", count)
			breaker.MarkFailed()
			// return
		}
		// failed
		_, err := faildFunc(context.Background(), nil)
		if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
	}
}

func Test_HalfFailedLoop(t *testing.T) {
	allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "Hello valid", nil
	}
	faildFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.InternalServer("", "")
	}
	fmt.Println("Start loop")
	breaker := sre.NewBreaker()
	count := 0
	// 50% failed
	for i := 0; i < 1000; i++ {
		// failed
		if err := breaker.Allow(); err != nil {
			// rejected
			// NOTE: when client reject requets locally,
			// continue add counter let the drop ratio higher.
			fmt.Println("Error: ", err)
			count++
			fmt.Println("count: ", count)
			breaker.MarkFailed()
			// return
		}
		// failed
		_, err := faildFunc(context.Background(), nil)
		if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
		// allowed
		if err := breaker.Allow(); err != nil {
			// rejected
			// NOTE: when client reject requets locally,
			// continue add counter let the drop ratio higher.
			fmt.Println("Error: ", err)
			count++
			fmt.Println("count: ", count)
			breaker.MarkFailed()
			// return
		}
		_, err = allowedFunc(context.Background(), nil)
		if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
	}
}

func Test_FailedServerWithoutGroup(t *testing.T) {
	// allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
	// 	return "Hello valid", nil
	// }
	faildFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.InternalServer("", "")
	}
	breaker := sre.NewBreaker()
	if err := breaker.Allow(); err != nil {
		// rejected
		// NOTE: when client reject requets locally,
		// continue add counter let the drop ratio higher.
		fmt.Println("Error: ", err)
		breaker.MarkFailed()
		// return
	}
	_, err := faildFunc(context.Background(), nil)
	if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
		breaker.MarkFailed()
	} else {
		breaker.MarkSuccess()
	}
}
