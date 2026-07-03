package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func getSMPTConfig() SMTPConfig {
	err := godotenv.Load()
	if err != nil {
		logrus.Info("⚠️  Warning: .env file not found, using system env")
	}
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587 // default
	}

	return SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

func HandleSendEmailTask(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	smtpConfig := getSMPTConfig()
	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.From)
	m.SetHeader("To", p.To)
	m.SetHeader("Subject", p.Subject)
	m.SetBody("text/html", p.Body)
	dialer := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password)
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func NewEmailDeliveryTask(To string, Subject string, Body string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailPayload{To: To, Subject: Subject, Body: Body})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("email:send", payload), nil
}

type CampaignEmailPayload struct {
	CampaignID  string `json:"campaign_id"`
	RecipientID string `json:"recipient_id"`
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
}

func NewCampaignEmailTask(campaignID, recipientID, to, subject, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(CampaignEmailPayload{
		CampaignID:  campaignID,
		RecipientID: recipientID,
		To:          to,
		Subject:     subject,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("email:campaign", payload), nil
}
