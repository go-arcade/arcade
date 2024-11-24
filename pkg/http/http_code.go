package http

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:25
 * @file: http_code.go
 * @description:
 */

var (
	StatusMovedPermanently = failed(301, "Moved Permanently")

	Failed = failed(500, "Request failed")

	// Unauthorized 401 sso
	Unauthorized           = failed(4401, "Unauthorized")
	AuthorizationIncorrect = failed(4402, "The sso format in the request header is incorrect")
	AuthorizationEmpty     = failed(4403, "Authorization is empty")
	InvalidToken           = failed(4404, "Invalid token")
	TokenBeEmpty           = failed(4407, "Token cannot be empty")
	TokenExpired           = failed(4408, "Token is expired")
	TokenFormatIncorrect   = failed(4409, "Token format is incorrect")

	InternalError = failed(5000, "Internal error, please contact the administrator")

	UserNotExist          = failed(4041, "User does not exist")
	UserAlreadyExist      = failed(4042, "User already exists")
	UserIncorrectPassword = failed(4043, "User incorrect password")

	UnsupportedProviders          = failed(4501, "Unsupported provider")
	ProviderIsRequired            = failed(4502, "Provider is required")
	InvalidStatusParameter        = failed(4502, "Invalid status parameter")
	TokenExchangeFailed           = failed(4503, "Token exchange failed")
	FailedToObtainUserInformation = failed(4504, "Failed to obtain user information")
)

var (
	Success = success(200, "Request Success")
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
