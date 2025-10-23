package middleware

import (
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/gofiber/fiber/v2"
)

// UnifiedResponse 统一响应
const (
	// DETAIL Detail 用于设置响应数据，例如查询，分页等，需要返回数据
	// e.g: c.Set(DETAIL, value)
	DETAIL = "detail"

	// OPERATION Operation 用于设置响应数据，例如新增，修改，删除等，不需要返回数据，只返回操作结果
	// e.g: c.Set(OPERATION, "")
	OPERATION = "operation"
)

// UnifiedResponseInterceptor 统一响应拦截器
// c.Locals("detail", value) 用于设置响应数据
// 如有其他需要，可自行添加
func UnifiedResponseMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			return err
		}

		// 业务逻辑错误
		if c.Response().StatusCode() != fiber.StatusOK {
			return httpx.WithRepErrMsg(c, httpx.Failed.Code, httpx.Failed.Msg, c.Path())
		}

		// 如果未设置响应状态码，默认将状态码设置为200（OK）
		if c.Response().StatusCode() == 0 {
			c.Status(fiber.StatusOK)
		}

		// 业务逻辑正确, 设置响应数据
		if c.Response().StatusCode() >= fiber.StatusOK && c.Response().StatusCode() < fiber.StatusMultipleChoices {
			if detail := c.Locals(DETAIL); detail != nil {
				return httpx.WithRepJSON(c, detail)
			}

			// 业务逻辑正确, 无响应数据, 只返回结果
			if c.Locals(OPERATION) != nil {
				return httpx.WithRepNotDetail(c)
			}
		}

		return nil
	}
}
