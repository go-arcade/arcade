package tool

import (
	"github.com/gofiber/fiber/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/22 23:03
 * @file: http.go
 * @description: http tool
 */

// GetLocalizedMessage 获取本地化消息
func GetLocalizedMessage(c *fiber.Ctx, messageId string, templateData map[string]string) string {
	// TODO: 实现Fiber的本地化消息获取
	return messageId
}

// GetLocalized 获取本地化消息, 不带模板数据
func GetLocalized(c *fiber.Ctx, messageId string) string {
	// TODO: 实现Fiber的本地化消息获取
	return messageId
}
