package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ConnectWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	userID := c.MustGet("user_id").(string)
	AppHub.Clients[userID] = conn
	defer func() {
		delete(AppHub.Clients, userID)
	}()
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		log.Printf("Received: %s", message)

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}
