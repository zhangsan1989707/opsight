package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

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

// SendEmail sends an email via SMTP.
func SendEmail(ch model.NotificationChannel, subject, body string) error {
	// Parse channel config for recipients
	var cfg struct {
		Recipients []string `json:"recipients"`
	}
	json.Unmarshal([]byte(ch.Config), &cfg)

	smtpHost := envOrDefault("SMTP_HOST", "")
	smtpPort := envOrDefault("SMTP_PORT", "587")
	smtpUser := envOrDefault("SMTP_USER", "")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpFrom := envOrDefault("SMTP_FROM", smtpUser)

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured (set SMTP_HOST, SMTP_USER, SMTP_PASSWORD)")
	}
	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("no recipients configured for channel %s", ch.Name)
	}

	addr := smtpHost + ":" + smtpPort
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	to := strings.Join(cfg.Recipients, ", ")
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		smtpFrom, to, subject, body)

	if err := smtp.SendMail(addr, auth, smtpFrom, cfg.Recipients, []byte(msg)); err != nil {
		return fmt.Errorf("SMTP send failed: %w", err)
	}

	logger.Info().Str("host", smtpHost).Strs("to", cfg.Recipients).Str("subject", subject).Msg("Email sent")
	return nil
}

// SendWeChatWork sends a markdown message via WeChat Work webhook.
func SendWeChatWork(ch model.NotificationChannel, title, content string) error {
	webhookURL := os.Getenv("WECHAT_WEBHOOK_URL")
	if webhookURL == "" {
		var cfg struct {
			WebhookURL string `json:"webhook_url"`
		}
		json.Unmarshal([]byte(ch.Config), &cfg)
		webhookURL = cfg.WebhookURL
	}

	if webhookURL == "" {
		return fmt.Errorf("WeChat webhook URL not configured")
	}

	// WeChat Work webhook expects: {"msgtype": "markdown", "markdown": {"content": "..."}}
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": fmt.Sprintf("### %s\n%s\n\n> 时间: %s", title, content, time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("WeChat webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeChat webhook returned status %d", resp.StatusCode)
	}

	// Verify response body
	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Errcode != 0 {
		return fmt.Errorf("WeChat webhook error: %d %s", result.Errcode, result.Errmsg)
	}

	logger.Info().Str("title", title).Msg("WeChat Work notification sent")
	return nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
