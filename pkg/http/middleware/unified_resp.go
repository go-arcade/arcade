package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/constant"
	httpx "github.com/observabil/arcade/pkg/http"
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
			if detail := c.Locals(constant.DETAIL); detail != nil {
				return httpx.WithRepJSON(c, detail)
			}

			// 业务逻辑正确, 无响应数据, 只返回结果
			if c.Locals(constant.OPERATION) != nil {
				return httpx.WithRepNotDetail(c)
			}
		}

		return nil
	}
}
