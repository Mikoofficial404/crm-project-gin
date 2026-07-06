package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

func NewHandleCampaignWhatsAppTask(
	sendWA func(phone, body string) error,
	updateStatus func(recipientID, status string, errMsg *string) error,
	markDoneIfFinished func(campaignID string) error,
) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p CampaignWhatsAppPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		logrus.Infof("Campaign WhatsApp task: campaignID=%s recipientID=%s phone=%s", p.CampaignID, p.RecipientID, p.Phone)

		sendErr := sendWA(p.Phone, p.Body)
		if sendErr != nil {
			errMsg := sendErr.Error()
			logrus.Warnf("Campaign WA gagal ke %s: %v", p.Phone, sendErr)
			updateStatus(p.RecipientID, "FAILED", &errMsg)
			return nil
		}

		updateStatus(p.RecipientID, "SENT", nil)
		markDoneIfFinished(p.CampaignID)
		return nil
	}
}

func NewHandleCampaignEmailTask(
	updateStatus func(recipientID, status string, errMsg *string) error,
	markDoneIfFinished func(campaignID string) error,
) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p CampaignEmailPayload
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
		sendErr := dialer.DialAndSend(m)

		if sendErr != nil {
			errMsg := sendErr.Error()
			logrus.Warnf("Campaign email gagal ke %s: %v", p.To, sendErr)
			updateStatus(p.RecipientID, "FAILED", &errMsg)
			return nil
		}

		updateStatus(p.RecipientID, "SENT", nil)
		markDoneIfFinished(p.CampaignID)
		return nil
	}
}
