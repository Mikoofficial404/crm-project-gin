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

func wrapHTMLTemplate(content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Inter', 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f4f4f5; margin: 0; padding: 0; color: #333333; }
        .container { max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05); }
        .header { background: linear-gradient(135deg, #2b7fff 0%%, #1e5bb8 100%%); padding: 32px 24px; text-align: center; }
        .header h1 { margin: 0; color: #ffffff; font-size: 26px; font-weight: 700; letter-spacing: -0.5px; }
        .content { padding: 40px 32px; line-height: 1.7; font-size: 16px; color: #374151; }
        .content p { margin-top: 0; margin-bottom: 20px; }
        .content a { color: #2b7fff; text-decoration: none; font-weight: 600; }
        .content a:hover { text-decoration: underline; }
        .footer { background-color: #f9fafb; padding: 24px; text-align: center; font-size: 13px; color: #6b7280; border-top: 1px solid #e5e7eb; }
        .footer p { margin: 0 0 8px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>CRM Workspace</h1>
        </div>
        <div class="content">
            %s
        </div>
        <div class="footer">
            <p>&copy; 2026 CRM Workspace. Hak Cipta Dilindungi.</p>
            <p>Pesan ini dikirim secara otomatis. Mohon tidak membalas email ini.</p>
        </div>
    </div>
</body>
</html>`, content)
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
	m.SetBody("text/html", wrapHTMLTemplate(p.Body))
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

type CampaignWhatsAppPayload struct {
	CampaignID  string `json:"campaign_id"`
	RecipientID string `json:"recipient_id"`
	Phone       string `json:"phone"`
	Body        string `json:"body"`
}

func NewCampaignWhatsAppTask(campaignID, recipientID, phone, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(CampaignWhatsAppPayload{
		CampaignID:  campaignID,
		RecipientID: recipientID,
		Phone:       phone,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("campaign:whatsapp", payload), nil
}
