// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
)

func AccessLogFormat(httpConfig *Http) fiber.Handler {
	// exclude api path
	// tips: 这里的路径是不需要记录日志的路径，url为端口后的全部路径
	excludedPaths := map[string]bool{
		"/health": true,
	}

	return func(c *fiber.Ctx) error {
		if httpConfig != nil && !httpConfig.AccessLog {
			return c.Next()
		}

		// 检查是否需要跳过日志
		if excludedPaths[c.Path()] {
			return c.Next()
		}

		// 记录开始时间
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 构建查询字符串
		query := c.Context().QueryArgs().String()
		queryStr := ""
		if query != "" {
			queryStr = "?" + query
		}

		body := c.Body()
		bodyStr := ""
		if body != nil {
			bodyStr = string(body)
		}

		log.Infow("HTTP request",
			"method", c.Method(),
			"path", c.Path(),
			"query", queryStr,
			"body", bodyStr,
			"latency", latency.String(),
			"status", c.Response().StatusCode(),
			"ip", c.IP(),
			"user-agent", c.Get("User-Agent"),
		)

		return err
	}
}
