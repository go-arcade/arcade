package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:24
 * @file: http_rep.go
 * @description:
 */

type Response struct {
	Code   int         `json:"code"`
	Detail interface{} `json:"detail,omitempty"`
	Msg    string      `json:"msg"`
}

// WithRepJSON 只返回json数据
func WithRepJSON(c *gin.Context, detail interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:   Success.Code,
		Detail: detail,
		Msg:    Success.Msg,
	})
}

// WithRepMsg 返回自定义code, msg
func WithRepMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// WithRepDetail 返回自定义code, msg, detail
func WithRepDetail(c *gin.Context, code int, msg string, detail interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:   code,
		Detail: detail,
		Msg:    msg,
	})
}

// WithRepNotDetail 只成功的返回操作结果，返回结构体没有detail字段
func WithRepNotDetail(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: Success.Code,
		Msg:  Success.Msg,
	})
}
