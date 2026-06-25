package websocket

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const gowaWSURL = "wss://gowa-83.semutssh.app/ws?device_id=23cea922-d3de-43c2-b359-12a96c7b4733"

func StartGowaWatcher() {
	for {
		err := connectAndListen()
		if err != nil {
			log.Printf("[GowaWatcher] Disconnected: %v — reconnecting in 5s...\n", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connectAndListen() error {
	log.Println("[GowaWatcher] Connecting to Gowa WebSocket...")

	credentials := base64.StdEncoding.EncodeToString([]byte("user991uia:pass2le4jt"))
	headers := http.Header{}
	headers.Set("Authorization", "Basic "+credentials)

	conn, _, err := websocket.DefaultDialer.Dial(gowaWSURL, headers)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	log.Println("[GowaWatcher]Connected! Listening for WhatsApp messages...")

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		switch messageType {
		case websocket.TextMessage:
			fmt.Printf("[GowaWatcher]Text Message: %s\n", string(payload))
		case websocket.BinaryMessage:
			fmt.Printf("[GowaWatcher]Binary Message (%d bytes)\n", len(payload))
		default:
			fmt.Printf("[GowaWatcher]Unknown message type %d\n", messageType)
		}
	}
}
