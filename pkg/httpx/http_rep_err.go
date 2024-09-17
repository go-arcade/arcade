package httpx

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:25
 * @file: http_rep_err.go
 * @description:
 */

type ResponseErr struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Err  any    `json:"err,omitempty"`
	Path string `json:"path,omitempty"`
}

// WithRepErr 只返回json数据
func WithRepErr(c *gin.Context, code int, msg string, err any, path string) {
	c.JSON(http.StatusOK, ResponseErr{
		Code: code,
		Msg:  msg,
		Err:  err,
		Path: path,
	})
}

// WithRepErrMsg 只返回json数据
func WithRepErrMsg(c *gin.Context, code int, msg string, path string) {
	c.JSON(http.StatusOK, ResponseErr{
		Code: code,
		Msg:  msg,
		Path: path,
	})
}

// WithRepErrNotData 只失败的返回操作结果，返回结构体没有path字段
func WithRepErrNotData(c *gin.Context, err string) {
	c.JSON(http.StatusOK, ResponseErr{
		Code: Success.Code,
		Msg:  Success.Msg,
		Err:  err,
	})
}
