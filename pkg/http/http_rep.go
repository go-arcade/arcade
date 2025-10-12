package http

import (
	"github.com/gofiber/fiber/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:24
 * @file: http_rep.go
 * @description:
 */

type Response struct {
	Code   int    `json:"code"`
	Detail any    `json:"detail,omitempty"`
	Msg    string `json:"msg"`
}

// WithRepJSON 只返回json数据
func WithRepJSON(c *fiber.Ctx, detail any) error {
	return c.JSON(Response{
		Code:   Success.Code,
		Detail: detail,
		Msg:    Success.Msg,
	})
}

// WithRepMsg 返回自定义code, msg
func WithRepMsg(c *fiber.Ctx, code int, msg string) error {
	return c.JSON(Response{
		Code: code,
		Msg:  msg,
	})
}

// WithRepDetail 返回自定义code, msg, detail
func WithRepDetail(c *fiber.Ctx, code int, msg string, detail any) error {
	return c.JSON(Response{
		Code:   code,
		Detail: detail,
		Msg:    msg,
	})
}

// WithRepNotDetail 只成功的返回操作结果，返回结构体没有detail字段
func WithRepNotDetail(c *fiber.Ctx) error {
	return c.JSON(Response{
		Code: Success.Code,
		Msg:  Success.Msg,
	})
}
