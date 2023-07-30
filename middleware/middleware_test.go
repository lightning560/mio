package middleware

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

var i int

func TestChain(t *testing.T) {
	next := func(ctx context.Context, req interface{}) (interface{}, error) { // biz_handler
		t.Log(req)
		i += 10
		fmt.Println("next")
		return "reply", nil
	}
	// Chain() return Middleware,(next) is handler type,next is params of middleware;(ctx,"helle mio!") is params of Handler;
	got, err := Chain(test1Middleware, test2Middleware, test3Middleware)(next)(context.Background(), "hello mio!")
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	if !reflect.DeepEqual(got, "reply") {
		t.Errorf("expect %v, got %v", "reply", got)
	}
	if !reflect.DeepEqual(i, 16) {
		t.Errorf("expect %v, got %v", 16, i)
	}
}

func test1Middleware(handler Handler) Handler {
	return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
		fmt.Println("test1 before")
		fmt.Println(req)
		i++
		reply, err = handler(ctx, req) //handler is test2Middleware return
		fmt.Println("test1 after")
		return
	}
}

func test2Middleware(handler Handler) Handler {
	return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
		fmt.Println("test2 before")
		fmt.Println(req)
		i += 2
		reply, err = handler(ctx, req) //handler is test3Middleware return
		fmt.Println("test2 after")
		return
	}
}

func test3Middleware(handler Handler) Handler {
	return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
		fmt.Println("test3 before")
		i += 3
		reply, err = handler(ctx, req) // handler is next
		fmt.Println("test3 after")
		return
	}
}
