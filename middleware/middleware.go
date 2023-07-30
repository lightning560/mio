package middleware

import "context"

type Handler func(ctx context.Context, req interface{}) (interface{}, error)

// 可以处理req 和 reply,核心handler只执行一次
type Middleware func(Handler) Handler

func Chain(m ...Middleware) Middleware {
	return func(h Handler) Handler { // return Middleware type
		for i := len(m) - 1; i >= 0; i-- { // 倒序
			h = m[i](h) // m[i] is Middleware,(h) is params,so return Handler type
			//  return m (h){mh3=m3(h), mh2= m2(mh3), mh1 = m1(mh2), return mh1} (params: h = biz_handler) return mh1(ctx,request)
			// 相当于三明治夹层
			// request 从mh1进入，mh1是最外层
		}
		return h
	}
}
