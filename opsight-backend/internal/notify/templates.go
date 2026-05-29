package notify

import (
	"fmt"
	"time"
)

// FormatAlertMessage returns a formatted title and body for an alert notification.
func FormatAlertMessage(ruleName, service, severity, condition, threshold, currentValue, timestamp string) (string, string) {
	tmpl := getSeverityTemplate(severity)
	title := fmt.Sprintf("[%s] %s - %s", severityEmoji(severity), severityLabel(severity), ruleName)
	body := fmt.Sprintf(tmpl, ruleName, service, condition, threshold, currentValue, timestamp)
	return title, body
}

func getSeverityTemplate(severity string) string {
	switch severity {
	case "critical":
		return `🚨 CRITICAL ALERT 🚨

Rule:     %s
Service:  %s
Condition: %s
Threshold: %s
Current:  %s
Time:     %s

ACTION REQUIRED: Immediate investigation needed. This is a critical issue that may affect service availability.
`
	case "warning":
		return `⚠️ WARNING ALERT ⚠️

Rule:     %s
Service:  %s
Condition: %s
Threshold: %s
Current:  %s
Time:     %s

Attention: This warning indicates a potential issue. Please review at your earliest convenience.
`
	case "info":
		return `ℹ️ INFO ALERT ℹ️

Rule:     %s
Service:  %s
Condition: %s
Threshold: %s
Current:  %s
Time:     %s

For your information only. No immediate action required.
`
	default:
		return `Alert: %s

Service:  %s
Condition: %s
Threshold: %s
Current:  %s
Time:     %s
`
	}
}

func severityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "⚪"
	}
}

func severityLabel(severity string) string {
	switch severity {
	case "critical":
		return "CRITICAL"
	case "warning":
		return "WARNING"
	case "info":
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// FormatTestMessage returns a simple test notification message.
func FormatTestMessage(channelName string) (string, string) {
	title := "[TEST] Notification Channel Test"
	body := fmt.Sprintf(`This is a test notification from Opsight.

Channel: %s
Time:    %s

If you received this message, the notification channel is configured correctly.`, channelName, time.Now().Format(time.RFC3339))
	return title, body
}
