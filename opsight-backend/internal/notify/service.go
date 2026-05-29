package notify

import (
	"encoding/json"
	"fmt"
	"os"

	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"
)

// SendAlert sends an alert notification via the specified channel asynchronously.
func SendAlert(channelID uint, ruleName string, severity string, title string, content string) {
	go func() {
		var ch model.NotificationChannel
		if err := database.DB.First(&ch, channelID).Error; err != nil {
			logger.Error().Err(err).Uint("channel_id", channelID).Msg("Failed to find notification channel")
			return
		}

		if !ch.Enabled {
			logger.Debug().Str("channel", ch.Name).Msg("Notification channel disabled, skipping")
			return
		}

		var sendErr error
		switch ch.Type {
		case "email":
			sendErr = SendEmail(ch, title, content)
		case "wechat_work":
			sendErr = SendWeChatWork(ch, title, content)
		default:
			sendErr = fmt.Errorf("unsupported channel type: %s", ch.Type)
		}

		status := "success"
		errStr := ""
		if sendErr != nil {
			status = "failed"
			errStr = sendErr.Error()
			logger.Error().Err(sendErr).Str("channel", ch.Name).Str("rule", ruleName).Msg("Notification send failed")
		} else {
			logger.Info().Str("channel", ch.Name).Str("rule", ruleName).Str("severity", severity).Msg("Notification sent successfully")
		}

		// Record history
		database.DB.Create(&model.NotificationHistory{
			ChannelID:   ch.ID,
			ChannelName: ch.Name,
			AlertRuleID: ruleName,
			Severity:    severity,
			Title:       title,
			Content:     content,
			Status:      status,
			Error:       errStr,
		})
	}()
}

// SendEmail is a stub that logs what would be sent via email.
func SendEmail(ch model.NotificationChannel, subject, body string) error {
	var cfg struct {
		SMTPHost string `json:"smtp_host"`
		SMTPPort string `json:"smtp_port"`
	}
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		logger.Warn().Err(err).Str("channel", ch.Name).Msg("Failed to parse email channel config")
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpHost == "" {
		smtpHost = "smtp.example.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if smtpFrom == "" {
		smtpFrom = "opsight@example.com"
	}

	logger.Info().
		Str("channel", ch.Name).
		Str("smtp_host", smtpHost).
		Str("smtp_port", smtpPort).
		Str("smtp_user", smtpUser).
		Str("smtp_from", smtpFrom).
		Str("subject", subject).
		Msg("EMAIL STUB - would send email notification")

	return nil
}

// SendWeChatWork is a stub that logs what would be sent via WeChat Work webhook.
func SendWeChatWork(ch model.NotificationChannel, title, content string) error {
	webhookURL := os.Getenv("WECHAT_WEBHOOK_URL")
	if webhookURL == "" {
		// Fallback to config from channel
		var cfg struct {
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
			logger.Warn().Err(err).Str("channel", ch.Name).Msg("Failed to parse wechat channel config")
		}
		webhookURL = cfg.WebhookURL
	}

	logger.Info().
		Str("channel", ch.Name).
		Str("webhook_url", webhookURL).
		Str("title", title).
		Msg("WECHAT_WORK STUB - would send WeChat Work notification")

	return nil
}
