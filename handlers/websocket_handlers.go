package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (optional)
	},
}

func WebSocketHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := c.GetString("role")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	client := &Client{
		UserID: userID,
		Role:   role,
		Conn:   conn,
	}

	Hub.Register(client)

	defer func() {
		Hub.Unregister(userID)
		conn.Close()
	}()

	// Just keep it alive — no reads for now
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}
