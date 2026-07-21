package websocket

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func StartGowaWatcher() {
	for {
		err := connectAndListen()
		if err != nil {
			logrus.WithError(err).Warn("[GowaWatcher] Disconnected, reconnecting in 5s...")
			time.Sleep(5 * time.Second)
		}
	}
}

func connectAndListen() error {
	logrus.Info("[GowaWatcher] Connecting to Gowa WebSocket...")

	wsURL := fmt.Sprintf("wss://%s/ws?device_id=%s", os.Getenv("WA_GOWA_URL")[8:], os.Getenv("WA_DEVICE_ID"))
	credentials := base64.StdEncoding.EncodeToString([]byte(os.Getenv("WA_BASIC_AUTH")))
	headers := http.Header{}
	headers.Set("Authorization", "Basic "+credentials)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	logrus.Info("[GowaWatcher] Connected! Listening for WhatsApp messages...")

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		switch messageType {
		case websocket.TextMessage:
			logrus.WithField("payload", string(payload)).Debug("[GowaWatcher] Text Message received")
			go func(data []byte) {
				appURL := os.Getenv("APP_URL")
				if appURL == "" {
					appURL = "http://localhost:8080"
				}
				_, err := http.Post(appURL+"/api/v1/webhook/whatsapp", "application/json", bytes.NewBuffer(data))
				if err != nil {
					logrus.WithError(err).Error("[GowaWatcher] Error forwarding to webhook")
				}
			}(payload)
		case websocket.BinaryMessage:
			logrus.WithField("bytes", len(payload)).Debug("[GowaWatcher] Binary Message received")
		default:
			logrus.WithField("type", messageType).Warn("[GowaWatcher] Unknown message type")
		}
	}
}
