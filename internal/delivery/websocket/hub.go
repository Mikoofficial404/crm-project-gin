package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients map[string]*websocket.Conn
}

var AppHub = &Hub{Clients: make(map[string]*websocket.Conn)}

func SendMessageToUser(userID string, message string) error {

	cable, ok := AppHub.Clients[userID]
	if !ok {
		return fmt.Errorf("user %s tidak ditemukan", userID)
	}

	cable.WriteMessage(websocket.TextMessage, []byte(message))
	return nil
}

func BroadcastMessage(message string) {
	for _, conn := range AppHub.Clients {
		conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
