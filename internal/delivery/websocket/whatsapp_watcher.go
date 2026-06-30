package websocket

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

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

	wsURL := fmt.Sprintf("wss://%s/ws?device_id=%s", os.Getenv("WA_GOWA_URL")[8:], os.Getenv("WA_DEVICE_ID"))
	credentials := base64.StdEncoding.EncodeToString([]byte(os.Getenv("WA_BASIC_AUTH")))
	headers := http.Header{}
	headers.Set("Authorization", "Basic "+credentials)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
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
			go func(data []byte) {
				_, err := http.Post("http://localhost:8080/api/v1/webhook/whatsapp", "application/json", bytes.NewBuffer(data))
				if err != nil {
					log.Printf("[GowaWatcher] Error forwarding to webhook: %v\n", err)
				}
			}(payload)
		case websocket.BinaryMessage:
			fmt.Printf("[GowaWatcher]Binary Message (%d bytes)\n", len(payload))
		default:
			fmt.Printf("[GowaWatcher]Unknown message type %d\n", messageType)
		}
	}
}
