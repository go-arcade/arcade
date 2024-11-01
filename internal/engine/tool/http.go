package tool

import (
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/22 23:03
 * @file: http.go
 * @description: http tool
 */

// GetLocalizedMessage 获取本地化消息
func GetLocalizedMessage(c *gin.Context, messageId string, templateData map[string]string) string {
	return ginI18n.MustGetMessage(c, &i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: templateData,
	})
}

// GetLocalized 获取本地化消息, 不带模板数据
func GetLocalized(c *gin.Context, messageId string) string {
	return ginI18n.MustGetMessage(c, &i18n.LocalizeConfig{
		MessageID: messageId,
	})
}
