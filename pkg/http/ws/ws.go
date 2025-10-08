package ws

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 23:53
 * @file: ws.go
 * @description: websocket
 */

type MessageType string

const (
	Heartbeat MessageType = "heartbeat"
	Log       MessageType = "log"
)

type Message struct {
	// message type
	Type   MessageType
	Detail any
}

func Handle(c *fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		defer func(c *websocket.Conn) {
			err := c.Close()
			if err != nil {
				log.Errorf("close websocket connection error: %v", err)
			}
		}(c)

		for {
			messageType, p, err := c.ReadMessage()
			if err != nil {
				break
			}

			var msg Message
			err = json.Unmarshal(p, &msg)
			if err != nil {
				log.Errorf("unmarshal message error: %v", err)
				break
			}

			switch msg.Type {
			case Heartbeat:
			// do something
			case Log:
			// do something
			default:
				log.Errorf("unknown message type")
			}

			err = c.WriteMessage(messageType, []byte("Received"))
			if err != nil {
				log.Errorf("write message error: %v", err)
				break
			}
		}
	})(c)
}
