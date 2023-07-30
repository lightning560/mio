package errors

import "net/http"

// All common ecode
var (
	OK         = Newf(http.StatusOK, "ok", "ok")                                  // 正确
	RequestErr = Newf(http.StatusBadRequest, "BadRequest", "params not validate") // 请求错误

	ErrServer             = Newf(http.StatusInternalServerError, "Internal Server Error", "server error")     // 服务器错误,503 http.StatusServiceUnavailable
	ErrNotFound           = Newf(http.StatusNotFound, "Not Found", "not found")                               // 未找到
	ErrUnauthorized       = Newf(http.StatusUnauthorized, "Unauthorized", "unauthorized")                     // 未认证
	ErrAccessDenied       = Newf(http.StatusForbidden, "Access Denied", "access denied")                      // 访问权限不足
	ErrServiceUnavailable = Newf(http.StatusServiceUnavailable, "Service Unavailable", "service unavailable") // 过载保护,服务暂不可用
	ErrDeadline           = Newf(http.StatusGatewayTimeout, "Gateway Timeout", "service timeout")             // 服务调用超时
	ErrMethodNotAllowed   = Newf(http.StatusMethodNotAllowed, "Method Not Allowed", "method not allowed")     // 不支持该方法
	ErrConflict           = Newf(http.StatusConflict, "Conflict", "conflict")                                 // 冲突
	ErrNothingFound       = Newf(http.StatusNotFound, "Nothing Found", "nothing found")                       // 啥都没有
	ErrFileNotExists      = Newf(http.StatusNotFound, "File Not Exists", "file not exists")                   // 上传文件不存在
	ErrFileTooLarge       = Newf(http.StatusRequestEntityTooLarge, "File Too Large", "file too large")        // 上传文件太大     // 登录失败次数太多
	// 用户不存在
	// code 403
	ErrPasswordTooLeak = Newf(http.StatusForbidden, "Password Too Leak", "password too leak")

	// code 400
	ErrInvalidParam = Newf(http.StatusBadRequest, "ErrInvalidParam", "invalid param") // 参数错误

	// ErrLimitExceed is exceed request limit.
	// code 429
	ErrLimitExceed = Newf(http.StatusTooManyRequests, "TooManyRequests", "service unavailable due to rate limit exceede") // 请求过多
	// ErrNotAllowed is request failed due to circuit breaker triggered.
	// code 503
	ErrNotAllowed = Newf(503, "circuit breaker", "request failed due to circuit breaker triggered") // 过载保护,服务暂不可用

	ErrNotLogin            = Newf(1010, "ErrNotLogin", "not login")
	ErrUserInactive        = Newf(1011, "ErrUserInactive", "user inactive")
	ErrUserNotExist        = Newf(1012, "User Not Exist", "user not exist")
	ErrAccessTokenExpires  = Newf(1020, "Access Token Expires", "Access Token Expires")   // Token 过期
	ErrPasswordHashExpires = Newf(1021, "PasswordHash Expires ", " PasswordHash Expires") // 密码时间戳过期

	ErrDegrade     = Newf(800, "server Degrade", "server Degrade")    // 被降级过滤的请求
	ErrRPCNoServer = Newf(801, "No RPC server", "invalid rpc server") // rpc服务的server都不可用
	ErrRPCNoClient = Newf(802, "No RPC Client", "invalid rpc client") // rpc服务的client都不可用
	ErrRPCNoAuth   = Newf(803, "RPCNoAuth", "RPCNoAuth")              // rpc服务的client没有授权

	//WarnLevel ecode,not critical
	// ErrNoDocuments is returned by the Query method when a query finds no documents in the result.
	ErrNoDocuments     = Newf(2001, "ErrNoDocuments", "no documents in result")  //
	ErrTargetNotExists = Newf(2002, "ErrFileNotExists", "file not exists")       //
	ErrAlreadyExists   = Newf(2003, "ErrAlreadyExists", "target already exists") //

	// AppKeyInvalid           = add(-1)   // 应用程序不存在或已被封禁
	// AccessKeyErr            = add(-2)   // Access Key错误
	// SignCheckErr            = add(-3)   // API校验密匙错误
	// MethodNoPermission      = add(-4)   // 调用方对该Method没有权限
	// NoLogin                 = add(-101) // 账号未登录
	// UserDisabled            = add(-102) // 账号被封停
	// LackOfScores            = add(-103) // 积分不足
	// LackOfCoins             = add(-104) // 硬币不足
	// CaptchaErr              = add(-105) // 验证码错误
	// UserInactive            = add(-106) // 账号未激活
	// UserNoMember            = add(-107) // 账号非正式会员或在适应期
	// AppDenied               = add(-108) // 应用不存在或者被封禁
	// MobileNoVerfiy          = add(-110) // 未绑定手机
	// CsrfNotMatchErr         = add(-111) // csrf 校验失败
	// ServiceUpdate           = add(-112) // 系统升级中
	// UserIDCheckInvalid      = add(-113) // 账号尚未实名认证
	// UserIDCheckInvalidPhone = add(-114) // 请先绑定手机
	// UserIDCheckInvalidCard  = add(-115) // 请先完成实名认证

	// NotModified           = add(-304) // 没有改动
	// TemporaryRedirect     = add(-307) // 撞车跳转

	// Unauthorized          = add(-401) // 未认证
	// AccessDenied          = add(-403) // 访问权限不足
	// NothingFound          = add(-404) // 啥都没有
	// MethodNotAllowed      = add(-405) // 不支持该方法
	// Conflict              = add(-409) // 冲突
	// ServerErr             = add(-500) // 服务器错误
	// ServiceUnavailable    = add(-503) // 过载保护,服务暂不可用
	// Deadline              = add(-504) // 服务调用超时
	// LimitExceed           = add(-509) // 超出限制
	// FileNotExists         = add(-616) // 上传文件不存在
	// FileTooLarge          = add(-617) // 上传文件太大
	// FailedTooManyTimes    = add(-625) // 登录失败次数太多
	// UserNotExist          = add(-626) // 用户不存在
	// PasswordTooLeak       = add(-628) // 密码太弱
	// UsernameOrPasswordErr = add(-629) // 用户名或密码错误
	// TargetNumberLimit     = add(-632) // 操作对象数量限制
	// TargetBlocked         = add(-643) // 被锁定
	// UserLevelLow          = add(-650) // 用户等级太低
	// UserDuplicate         = add(-652) // 重复的用户
	// AccessTokenExpires    = add(-658) // Token 过期
	// PasswordHashExpires   = add(-662) // 密码时间戳过期
	// AreaLimit             = add(-688) // 地理区域限制
	// CopyrightLimit        = add(-689) // 版权限制
	// FailToAddMoral        = add(-701) // 扣节操失败

	// Degrade     = add(-1200) // 被降级过滤的请求
	// RPCNoClient = add(-1201) // rpc服务的client都不可用
	// RPCNoAuth   = add(-1202) // rpc服务的client没有授权
)
