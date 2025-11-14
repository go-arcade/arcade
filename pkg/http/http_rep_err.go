package http

import (
	"github.com/gofiber/fiber/v2"
)


type ResponseErr struct {
	ErrCode int    `json:"code"`
	ErrMsg  any    `json:"errMsg"`
	Path    string `json:"path,omitempty"`
}

// WithRepErr 返回操作结果，返回结构体有path字段
func WithRepErr(c *fiber.Ctx, code int, errMsg string, path string) error {
	return c.JSON(ResponseErr{
		ErrCode: code,
		ErrMsg:  errMsg,
		Path:    path,
	})
}

// WithRepErrMsg 只返回json数据
func WithRepErrMsg(c *fiber.Ctx, code int, errMsg string, path string) error {
	return c.JSON(ResponseErr{
		ErrCode: code,
		ErrMsg:  errMsg,
		Path:    path,
	})
}

// WithRepErrNotData 只失败的返回操作结果，返回结构体没有path字段
func WithRepErrNotData(c *fiber.Ctx, errMsg string) error {
	return c.JSON(ResponseErr{
		ErrCode: Success.Code,
		ErrMsg:  errMsg,
	})
}
