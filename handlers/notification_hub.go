package handlers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Each connected client
type Client struct {
	UserID uint
	Role   string
	Conn   *websocket.Conn
}

// Hub to track connected clients
type NotificationHub struct {
	clients map[uint]*Client
	lock    sync.RWMutex
}

// Global Hub instance
var Hub = NotificationHub{
	clients: make(map[uint]*Client),
}

// Register new connection
func (h *NotificationHub) Register(c *Client) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Normalize role to uppercase
	c.Role = strings.ToUpper(c.Role)

	h.clients[c.UserID] = c
	fmt.Printf("✅ User %d connected via WebSocket (%s)\n", c.UserID, c.Role)
}

// Remove disconnected client
func (h *NotificationHub) Unregister(userID uint) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.clients, userID)
	fmt.Printf("❌ User %d disconnected\n", userID)
}

// Send notification to one user
func (h *NotificationHub) SendToUser(userID uint, message string) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if client, ok := h.clients[userID]; ok {
		client.Conn.WriteJSON(map[string]string{"message": message})
	}
}

// Broadcast to all superadmins
func (h *NotificationHub) BroadcastToSuperAdmins(message string) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	for _, client := range h.clients {
		// Case-insensitive check
		if strings.EqualFold(client.Role, "SUPERADMIN") {
			client.Conn.WriteJSON(map[string]string{"message": message})
		}
	}
}
