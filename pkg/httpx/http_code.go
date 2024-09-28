package httpx

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:25
 * @file: http_code.go
 * @description:
 */

var (
	Failed = failed(500, "request failed")

	Unauthorized           = failed(4001, "unauthorized")
	AuthorizationIncorrect = failed(4002, "The auth format in the request header is incorrect")
	AuthorizationEmpty     = failed(4003, "Authorization is empty")
	TokenInvalid           = failed(4004, "Token is invalid")
	TokenEmpty             = failed(405, "Token is empty")

	ErrUserPhone = failed(10001, "用户手机号不合法")
	ErrSignParam = failed(10002, "签名参数有误")

	InternalError = failed(500, "internal error, please contact the administrator")
)

var (
	Success = success(200, "success")
)

// failed 构造函数
func failed(code int, msg string) *Response {
	return &Response{
		Code:   code,
		Msg:    msg,
		Detail: nil,
	}
}

// success 构造函数
func success(code int, msg string) *Response {
	return &Response{
		Code:   code,
		Msg:    msg,
		Detail: nil,
	}
}
