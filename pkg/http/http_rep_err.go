package http

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
	ErrCode int    `json:"code"`
	ErrMsg  any    `json:"errMsg"`
	Path    string `json:"path,omitempty"`
}

// WithRepErr 返回操作结果，返回结构体有path字段
func WithRepErr(c *gin.Context, code int, errMsg string, path string) {
	c.JSON(http.StatusOK, ResponseErr{
		ErrCode: code,
		ErrMsg:  errMsg,
		Path:    path,
	})
}

// WithRepErrMsg 只返回json数据
func WithRepErrMsg(c *gin.Context, code int, errMsg string, path string) {
	c.JSON(http.StatusOK, ResponseErr{
		ErrCode: code,
		ErrMsg:  errMsg,
		Path:    path,
	})
}

// WithRepErrNotData 只失败的返回操作结果，返回结构体没有path字段
func WithRepErrNotData(c *gin.Context, errMsg string) {
	c.JSON(http.StatusOK, ResponseErr{
		ErrCode: Success.Code,
		ErrMsg:  errMsg,
	})
}
