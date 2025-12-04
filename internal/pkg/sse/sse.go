package sse

import (
	"sync"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

type SSEHub struct {
	mu      sync.RWMutex
	subs    map[string]map[chan []byte]struct{} // task_id -> set(ch)
	closing chan struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		subs:    make(map[string]map[chan []byte]struct{}),
		closing: make(chan struct{}),
	}
}
func (h *SSEHub) Broadcast(taskID string, payload any) {
	data, _ := sonic.Marshal(payload)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[taskID] {
		select {
		case ch <- data:
		default:
			// 慢消费者丢弃
		}
	}
}

func (h *SSEHub) Subscribe(taskID string) (ch chan []byte, cancel func()) {
	ch = make(chan []byte, 128)
	h.mu.Lock()
	if _, ok := h.subs[taskID]; !ok {
		h.subs[taskID] = make(map[chan []byte]struct{})
	}
	h.subs[taskID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if set, ok := h.subs[taskID]; ok {
			delete(set, ch)
			if len(set) == 0 {
				delete(h.subs, taskID)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
}
func (h *SSEHub) HTTPHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		taskID := c.Query("task_id")
		if taskID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("missing task_id")
		}

		ch, cancel := h.Subscribe(taskID)
		defer cancel()

		notify := c.Context().Done()
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return c.SendStatus(fiber.StatusInternalServerError)
				}
				c.WriteString(string(data))
			case <-notify:
				return c.SendStatus(fiber.StatusOK)
			}
		}
	}
}
