package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// LogStreamHandler WebSocket日志流处理器
type LogStreamHandler struct {
	logAggregator *LogAggregator
}

// NewLogStreamHandler 创建WebSocket日志流处理器
func NewLogStreamHandler(logAggregator *LogAggregator) *LogStreamHandler {
	return &LogStreamHandler{
		logAggregator: logAggregator,
	}
}

// LogStreamRequest WebSocket请求消息
type LogStreamRequest struct {
	Type     string `json:"type"`      // subscribe, unsubscribe
	TaskID   string `json:"task_id"`   // 任务ID
	FromLine int32  `json:"from_line"` // 从第几行开始
}

// LogStreamResponse WebSocket响应消息
type LogStreamResponse struct {
	Type    string    `json:"type"`    // log, error, complete
	TaskID  string    `json:"task_id"` // 任务ID
	Log     *LogEntry `json:"log,omitempty"`
	Error   string    `json:"error,omitempty"`
	Message string    `json:"message,omitempty"`
}

// Upgrade 创建WebSocket升级中间件
func (h *LogStreamHandler) Upgrade() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		defer func() {
			if err := conn.Close(); err != nil {
				log.Errorf("failed to close websocket: %v", err)
			}
		}()

		// 心跳ticker
		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		// 上下文和取消函数
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 活跃订阅
		var activeSubscription <-chan *LogEntry
		var currentTaskID string

		// 发送消息的辅助函数
		sendResponse := func(resp *LogStreamResponse) error {
			data, err := sonic.Marshal(resp)
			if err != nil {
				log.Errorf("failed to marshal response: %v", err)
				return err
			}
			return conn.WriteMessage(websocket.TextMessage, data)
		}

		// 启动接收消息的goroutine
		messageChan := make(chan *LogStreamRequest, 10)
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Infof("websocket connection closed: %v", err)
					cancel()
					return
				}

				var req LogStreamRequest
				if err := sonic.Unmarshal(message, &req); err != nil {
					log.Errorf("failed to unmarshal request: %v", err)
					sendResponse(&LogStreamResponse{
						Type:  "error",
						Error: fmt.Sprintf("invalid request format: %v", err),
					})
					continue
				}

				messageChan <- &req
			}
		}()

		// 主循环
		for {
			select {
			case <-ctx.Done():
				return

			case <-heartbeatTicker.C:
				// 发送心跳
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Errorf("failed to send ping: %v", err)
					return
				}

			case req := <-messageChan:
				switch req.Type {
				case "subscribe":
					// 处理订阅请求
					if currentTaskID != "" && currentTaskID != req.TaskID {
						// 取消之前的订阅
						cancel()
						activeSubscription = nil
					}

					currentTaskID = req.TaskID
					log.Infof("client subscribing to task %s from line %d", req.TaskID, req.FromLine)

					// 先发送历史日志
					go func() {
						historicalLogs, err := h.logAggregator.GetLogsByTaskID(req.TaskID, req.FromLine, 1000)
						if err != nil {
							log.Errorf("failed to get historical logs: %v", err)
							sendResponse(&LogStreamResponse{
								Type:   "error",
								TaskID: req.TaskID,
								Error:  fmt.Sprintf("failed to load history: %v", err),
							})
							return
						}

						// 发送历史日志
						for _, entry := range historicalLogs {
							if err := sendResponse(&LogStreamResponse{
								Type:   "log",
								TaskID: req.TaskID,
								Log:    entry,
							}); err != nil {
								log.Errorf("failed to send historical log: %v", err)
								return
							}
						}

						// 发送历史日志完成标记
						sendResponse(&LogStreamResponse{
							Type:    "message",
							TaskID:  req.TaskID,
							Message: "historical logs loaded",
						})
					}()

					// 订阅实时日志
					activeSubscription = h.logAggregator.Subscribe(ctx, req.TaskID)

					// 确认订阅成功
					sendResponse(&LogStreamResponse{
						Type:    "message",
						TaskID:  req.TaskID,
						Message: "subscribed",
					})

				case "unsubscribe":
					// 取消订阅
					if currentTaskID == req.TaskID {
						currentTaskID = ""
						activeSubscription = nil

						sendResponse(&LogStreamResponse{
							Type:    "message",
							TaskID:  req.TaskID,
							Message: "unsubscribed",
						})
					}

				default:
					sendResponse(&LogStreamResponse{
						Type:  "error",
						Error: fmt.Sprintf("unknown request type: %s", req.Type),
					})
				}

			case entry, ok := <-activeSubscription:
				if !ok {
					// channel关闭
					activeSubscription = nil
					if currentTaskID != "" {
						sendResponse(&LogStreamResponse{
							Type:    "complete",
							TaskID:  currentTaskID,
							Message: "log stream completed",
						})
					}
					continue
				}

				// 发送实时日志
				if err := sendResponse(&LogStreamResponse{
					Type:   "log",
					TaskID: currentTaskID,
					Log:    entry,
				}); err != nil {
					log.Errorf("failed to send real-time log: %v", err)
					return
				}
			}
		}
	})
}
