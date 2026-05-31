package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"opsight-backend/internal/database"
	"opsight-backend/internal/handler"
	"opsight-backend/internal/metrics"
	"opsight-backend/internal/model"
	"opsight-backend/internal/notify"
	"opsight-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// StartAlertEvaluator runs the alert evaluation loop every 30 seconds.
func StartAlertEvaluator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		UpdateServiceMetrics()
		evaluateRules()
	}
}

func evaluateRules() {
	var rules []model.AlertRule
	database.DB.Where("enabled = ?", true).Find(&rules)
	if len(rules) == 0 {
		return
	}

	var agents []model.AgentInstance
	database.DB.Where("status = ?", "online").Find(&agents)

	for _, rule := range rules {
		condition := strings.TrimSpace(rule.Condition)
		parts := strings.SplitN(condition, " > ", 2)
		if len(parts) != 2 {
			logger.Debug().Str("rule", rule.ID).Str("condition", condition).Msg("Rule condition unable to parse, skipping")
			continue
		}
		metricName := strings.TrimSpace(parts[0])
		thresholdStr := strings.TrimSpace(parts[1])

		threshold, err := strconv.ParseFloat(strings.TrimSuffix(thresholdStr, "%"), 64)
		if err != nil {
			logger.Debug().Str("rule", rule.ID).Str("threshold", thresholdStr).Msg("Rule threshold not numeric, skipping")
			continue
		}

		for _, agent := range agents {
			var latest model.MetricSnapshot
			if err := database.DB.Where("agent_id = ?", agent.ID).Order("reported_at DESC").First(&latest).Error; err != nil {
				continue
			}

			if time.Since(latest.ReportedAt) > 2*time.Minute {
				continue
			}

			var currentValue float64
			evaluatable := true

			switch {
			case strings.HasPrefix(metricName, "cpu_usage"):
				currentValue = latest.CPUPercent
			case strings.HasPrefix(metricName, "memory_usage"):
				currentValue = latest.MemPercent
			case strings.HasPrefix(metricName, "disk_usage"):
				currentValue = latest.DiskPercent
			case strings.HasPrefix(metricName, "p99"):
				currentValue = latest.Load1
			default:
				evaluatable = false
			}

			if !evaluatable {
				continue
			}

			var existingAlert model.AlertEvent
			alreadyFiring := database.DB.Where(
				"alert_rule_id = ? AND hostname = ? AND status = ?",
				rule.ID, agent.Hostname, "firing",
			).First(&existingAlert).Error == nil

			if currentValue > threshold {
				if !alreadyFiring {
					msg := fmt.Sprintf("%s: %.1f > %.1f", rule.Name, currentValue, threshold)
					alertEvent := model.AlertEvent{
						AlertRuleID: rule.ID,
						RuleName:    rule.Name,
						Hostname:    agent.Hostname,
						Severity:    rule.Severity,
						Message:     msg,
						MetricValue: currentValue,
						Threshold:   threshold,
						Status:      "firing",
					}
					database.DB.Create(&alertEvent)

					rule.LastTriggered = "just now"
					database.DB.Model(&rule).Update("last_triggered", rule.LastTriggered)

					sendNotificationForRule(rule, currentValue, threshold, agent.Hostname)

					logger.Info().
						Str("rule", rule.ID).
						Str("rule_name", rule.Name).
						Str("hostname", agent.Hostname).
						Float64("value", currentValue).
						Float64("threshold", threshold).
						Msg("Alert fired")

					handler.BroadcastEvent("alert_firing", gin.H{
						"rule_id":   rule.ID,
						"name":      rule.Name,
						"hostname":  agent.Hostname,
						"severity":  rule.Severity,
						"value":     currentValue,
						"threshold": threshold,
					})
					metrics.RecordAlertFired()
				}
			} else {
				if alreadyFiring {
					now := time.Now()
					database.DB.Model(&existingAlert).Updates(map[string]interface{}{
						"status":      "resolved",
						"resolved_at": now,
					})
					logger.Info().
						Str("rule", rule.ID).
						Str("hostname", agent.Hostname).
						Msg("Alert resolved")
					metrics.RecordAlertResolved()
				}
			}
		}
	}

	handler.BroadcastEvent("service_status", gin.H{"message": "Status update"})
}

func sendNotificationForRule(rule model.AlertRule, value, threshold float64, hostname string) {
	var channels []model.NotificationChannel
	database.DB.Where("enabled = ?", true).Find(&channels)
	if len(channels) == 0 {
		return
	}

	for _, ch := range channels {
		title, body := notify.FormatAlertMessage(
			rule.Name,
			rule.Service,
			rule.Severity,
			rule.Condition,
			fmt.Sprintf("%.1f", threshold),
			fmt.Sprintf("%.1f", value),
			time.Now().Format(time.RFC3339),
		)

		logger.Info().
			Str("channel", ch.Name).
			Str("rule", rule.Name).
			Str("hostname", hostname).
			Str("severity", rule.Severity).
			Msg("Alert firing - sending notification")

		notify.SendAlert(ch.ID, rule.Name, rule.Severity, title, body)
	}
}

// UpdateServiceMetrics is a placeholder for computing real service metrics from agent data.
// Previously wrote random demo data to the DB; this has been removed.
func UpdateServiceMetrics() {
	// TODO: aggregate real service metrics from MetricSnapshot data
}
