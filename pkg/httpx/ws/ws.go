package ws

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	httpx "github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gorilla/websocket"
	"net/http"
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
	Detail interface{}
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handle(r *gin.Context) {
	conn, err := upgrade.Upgrade(r.Writer, r.Request, nil)
	if err != nil {
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("close websocket connection error: %v", err)
		}
	}(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		err = json.Unmarshal(p, &msg)

		switch msg.Type {
		case Heartbeat:
		// do something
		case Log:
		// do something
		default:
			httpx.WithRepErrMsg(r, http.StatusBadRequest, "unknown message type", r.Request.URL.Path)
		}

		err = conn.WriteMessage(messageType, []byte("Received"))
		if err != nil {
			log.Errorf("write message error: %v", err)
			break
		}
	}
}
