package handler

import (
	"net/http"
	"sync"
	"time"

	"opsight-backend/internal/auth"
	"opsight-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return false }}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.RWMutex
)

// SetWSOriginCheck sets the WebSocket upgrader's origin check function.
// Must be called during startup after allowedOrigins are computed.
func SetWSOriginCheck(checkFunc func(r *http.Request) bool) {
	upgrader.CheckOrigin = checkFunc
}

func HandleWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	if _, err := auth.ValidateToken(token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error().Err(err).Msg("WebSocket upgrade error")
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastEvent sends an event to all connected WebSocket clients.
func BroadcastEvent(eventType string, data interface{}) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	msg := gin.H{"type": eventType, "data": data, "time": time.Now().UTC()}
	for conn := range clients {
		if err := conn.WriteJSON(msg); err != nil {
			logger.Warn().Err(err).Msg("WebSocket write error")
		}
	}
}
