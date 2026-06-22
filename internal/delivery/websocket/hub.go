package websocket

import (
	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients map[*websocket.Conn]string
}

func SendMessageToUser(userID string, message string) {

}
