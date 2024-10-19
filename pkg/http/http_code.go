package http

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
	InvalidToken           = failed(4004, "Invalid token")
	InValidAccessToken     = failed(4005, "Invalid access token")
	InValidRefreshToken    = failed(4006, "Invalid refresh token")
	TokenBeEmpty           = failed(405, "token cannot be empty")

	InternalError = failed(500, "internal error, please contact the administrator")

	UserNotExist          = failed(404, "user does not exist")
	UserAlreadyExist      = failed(405, "user already exists")
	UserIncorrectPassword = failed(406, "user incorrect password")
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
