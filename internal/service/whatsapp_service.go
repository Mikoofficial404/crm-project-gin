package service

import (
	"bytes"
	"crm-project/pkg/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type WhatsAppService struct {
	GowaURL    string
	DeviceID   string
	BasicAuth  string
	HTTPClient *http.Client
}

func NewWhatsAppService(gowaURL, deviceID, basicAuth string) *WhatsAppService {
	return &WhatsAppService{
		GowaURL:    gowaURL,
		DeviceID:   deviceID,
		BasicAuth:  basicAuth,
		HTTPClient: &http.Client{},
	}
}

func (s *WhatsAppService) SendWA(phone string, textMessage string) error {

	normalizedPhone := utils.NormalizePhone(phone)

	payload := map[string]interface{}{
		"device_id": s.DeviceID,
		"phone":     normalizedPhone,
		"message":   textMessage,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gagal JSON: %w", err)
	}

	endpoint := fmt.Sprintf("%s/send/message", s.GowaURL)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("gagal HTTP: %w", err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString([]byte(s.BasicAuth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Id", s.DeviceID)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Gowa err: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("gagal kirim WA (Status %d): %s", resp.StatusCode, string(body))
	}

	logrus.WithField("phone", phone).Info("[WhatsAppService] Berhasil mengirim WA ke Klien")
	return nil
}
