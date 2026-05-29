package audit

import (
	"os"
	"time"

	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"
)

// IsEnabled returns whether audit logging is enabled.
func IsEnabled() bool {
	return os.Getenv("AUDIT_ENABLED") != "false"
}

// Log creates an audit log entry asynchronously to avoid blocking the request.
func Log(userID uint, userName, action, resource, resourceID, detail, ip, userAgent, status string) {
	if !IsEnabled() {
		return
	}
	go LogAsync(userID, userName, action, resource, resourceID, detail, ip, userAgent, status)
}

// LogAsync writes an audit log entry to the database in a goroutine.
func LogAsync(userID uint, userName, action, resource, resourceID, detail, ip, userAgent, status string) {
	entry := model.AuditLog{
		UserID:     userID,
		UserName:   userName,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		UserAgent:  userAgent,
		Status:     status,
		CreatedAt:  time.Now(),
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		logger.Error().Err(err).
			Str("action", action).
			Str("resource", resource).
			Str("user_name", userName).
			Msg("Failed to write audit log")
	}
}

// Query returns paginated audit log entries with optional filters.
func Query(userID uint, action, resource string, startTime, endTime time.Time, page, pageSize int) ([]model.AuditLog, int64, error) {
	db := database.DB.Model(&model.AuditLog{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if resource != "" {
		db = db.Where("resource = ?", resource)
	}
	if !startTime.IsZero() {
		db = db.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		db = db.Where("created_at <= ?", endTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var logs []model.AuditLog
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ActionBreakdown represents a count of audit entries grouped by action.
type ActionBreakdown struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// Stats returns audit log statistics.
func Stats() (totalCount int64, todayCount int64, actionsBreakdown []ActionBreakdown, err error) {
	if err := database.DB.Model(&model.AuditLog{}).Count(&totalCount).Error; err != nil {
		return 0, 0, nil, err
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	if err := database.DB.Model(&model.AuditLog{}).Where("created_at >= ?", todayStart).Count(&todayCount).Error; err != nil {
		return 0, 0, nil, err
	}

	var breakdown []ActionBreakdown
	if err := database.DB.Model(&model.AuditLog{}).
		Select("action, count(*) as count").
		Group("action").
		Order("count DESC").
		Scan(&breakdown).Error; err != nil {
		return 0, 0, nil, err
	}

	return totalCount, todayCount, breakdown, nil
}
