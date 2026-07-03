package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

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
