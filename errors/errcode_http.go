package errors

// 40xx is prefix of client error,5xx is prefix of server error
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
var (
	ErrGinBind             = Newf(1401, "ErrGinBind", "body don't match proto,please check body&proto") //
	ErrHttpRequestParam    = Newf(1402, "Http Request Invalid Param", "invalid param")                  // 参数错误
	ErrHttpRequestValidate = Newf(1403, "Http Request Validate", "validate error.please check params")  //
)
