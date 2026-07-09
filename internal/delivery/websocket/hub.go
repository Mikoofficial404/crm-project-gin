package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients map[string]map[*websocket.Conn]bool
}

var AppHub = &Hub{Clients: make(map[string]map[*websocket.Conn]bool)}

func SendMessageToUser(userID string, message string) error {
	userConns, ok := AppHub.Clients[userID]
	if !ok || len(userConns) == 0 {
		return fmt.Errorf("user %s tidak memiliki koneksi aktif", userID)
	}

	for conn := range userConns {
		conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
	return nil
}

func BroadcastMessage(message string) {
	for _, userConns := range AppHub.Clients {
		for conn := range userConns {
			conn.WriteMessage(websocket.TextMessage, []byte(message))
		}
	}
}
