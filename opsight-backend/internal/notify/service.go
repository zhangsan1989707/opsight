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

	"github.com/rs/zerolog"
)

func SendAlert(channelID uint, ruleName string, severity string, title string, content string) {
	go func() {
		err := sendAlertWithRetry(channelID, ruleName, severity, title, content)
		recordNotification(channelID, ruleName, severity, title, content, err)
	}()
}

func sendAlertWithRetry(channelID uint, ruleName string, severity string, title string, content string) error {
	var ch model.NotificationChannel
	if err := database.DB.First(&ch, channelID).Error; err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	if !ch.Enabled {
		logger.Debug().Str("channel", ch.Name).Msg("Notification channel disabled, skipping")
		return nil
	}

	var lastErr error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = sendByChannel(ch, title, content)
		if lastErr == nil {
			logger.Info().
				Str("channel", ch.Name).
				Str("rule", ruleName).
				Str("severity", severity).
				Int("attempt", attempt).
				Msg("Notification sent successfully")
			return nil
		}

		logger.Warn().
			Err(lastErr).
			Str("channel", ch.Name).
			Str("rule", ruleName).
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Msg("Notification send failed, retrying with backoff")

		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
}

func sendByChannel(ch model.NotificationChannel, title, content string) error {
	switch ch.Type {
	case "email":
		return sendEmailWithTimeout(ch, title, content)
	case "wechat_work":
		return sendWeChatWorkWithTimeout(ch, title, content)
	default:
		return fmt.Errorf("unsupported channel type: %s", ch.Type)
	}
}

func recordNotification(channelID uint, ruleName string, severity string, title string, content string, sendErr error) {
	status := "success"
	errStr := ""
	if sendErr != nil {
		status = "failed"
		errStr = sendErr.Error()
		logger.Error().Err(sendErr).Str("channel_id", fmt.Sprintf("%d", channelID)).Str("rule", ruleName).Msg("Notification send failed after all retries")
	}

	database.DB.Create(&model.NotificationHistory{
		ChannelID:   channelID,
		ChannelName: "",
		AlertRuleID: ruleName,
		Severity:    severity,
		Title:       title,
		Content:     content,
		Status:      status,
		Error:       errStr,
	})
}

func sendEmailWithTimeout(ch model.NotificationChannel, subject, body string) error {
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

	err := smtp.SendMail(addr, auth, smtpFrom, cfg.Recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("SMTP send failed: %w", err)
	}

	logger.Info().Str("host", smtpHost).Strs("to", cfg.Recipients).Str("subject", subject).Msg("Email sent")
	return nil
}

func sendWeChatWorkWithTimeout(ch model.NotificationChannel, title, content string) error {
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("WeChat webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeChat webhook returned status %d", resp.StatusCode)
	}

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

func logAlert(evt *zerolog.Event, msg string) *zerolog.Event {
	return evt
}