package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetAlertRules(c *gin.Context) {
	var rules []model.AlertRule
	database.DB.Find(&rules)

	result := make([]dto.AlertRuleDTO, len(rules))
	for i, r := range rules {
		result[i] = dto.FromAlertRule(r)
	}
	response.Success(c, gin.H{"rules": result, "total": len(result)})
}

func ToggleAlertRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.AlertRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "rule not found")
		return
	}
	rule.Enabled = !rule.Enabled
	database.DB.Save(&rule)

	userID, email, _ := auth.GetCurrentUser(c)
	action := "enabled"
	if !rule.Enabled {
		action = "disabled"
	}
	audit.Log(userID, email, "toggle", "alert-rules", id, "Alert rule "+action+": "+rule.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dto.FromAlertRule(rule))
}

func ListAlertEvents(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	hostname := c.Query("hostname")

	db := database.DB.Model(&model.AlertEvent{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if severity != "" {
		db = db.Where("severity = ?", severity)
	}
	if hostname != "" {
		db = db.Where("hostname = ?", hostname)
	}

	var events []model.AlertEvent
	db.Order("created_at DESC").Find(&events)

	response.Success(c, gin.H{"events": events, "total": len(events)})
}

func GetAlertEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid event id")
		return
	}

	var event model.AlertEvent
	if err := database.DB.First(&event, uint(id)).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "alert event not found")
		return
	}

	response.Success(c, gin.H{"event": event})
}

func ResolveAlertEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid event id")
		return
	}

	var event model.AlertEvent
	if err := database.DB.First(&event, uint(id)).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "alert event not found")
		return
	}

	if event.Status == "resolved" {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "alert event already resolved")
		return
	}

	now := time.Now()
	database.DB.Model(&event).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": now,
	})

	response.Success(c, gin.H{"event": event})
}

func CreateAlertRule(c *gin.Context) {
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("RULE-%d", time.Now().UnixNano()%100000)
	}
	if err := database.DB.Create(&rule).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to create rule")
		return
	}

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "create", "alert-rules", rule.ID, "Created alert rule: "+rule.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dto.FromAlertRule(rule))
}

func UpdateAlertRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.AlertRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "rule not found")
		return
	}

	var update model.AlertRule
	if err := c.ShouldBindJSON(&update); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}

	database.DB.Model(&rule).Updates(model.AlertRule{
		Name:      update.Name,
		Condition: update.Condition,
		Threshold: update.Threshold,
		Service:   update.Service,
		Severity:  update.Severity,
		Enabled:   update.Enabled,
	})

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "update", "alert-rules", id, "Updated alert rule: "+rule.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	database.DB.Where("id = ?", id).First(&rule)
	response.Success(c, dto.FromAlertRule(rule))
}

func DeleteAlertRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.AlertRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "rule not found")
		return
	}

	database.DB.Delete(&rule)

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "delete", "alert-rules", id, "Deleted alert rule: "+rule.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, gin.H{"message": "rule deleted"})
}
