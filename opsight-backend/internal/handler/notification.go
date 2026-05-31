package handler

import (
	"net/http"
	"strconv"

	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/internal/notify"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func ListNotificationChannels(c *gin.Context) {
	var channels []model.NotificationChannel
	database.DB.Find(&channels)
	response.Success(c, gin.H{"channels": channels})
}

func CreateNotificationChannel(c *gin.Context) {
	var ch model.NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if err := database.DB.Create(&ch).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to create channel")
		return
	}
	response.Success(c, gin.H{"channel": ch})
}

func UpdateNotificationChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid channel id")
		return
	}

	var ch model.NotificationChannel
	if err := database.DB.First(&ch, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "channel not found")
		return
	}

	var update model.NotificationChannel
	if err := c.ShouldBindJSON(&update); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}

	database.DB.Model(&ch).Updates(update)
	response.Success(c, gin.H{"channel": ch})
}

func DeleteNotificationChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid channel id")
		return
	}

	var ch model.NotificationChannel
	if err := database.DB.First(&ch, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "channel not found")
		return
	}

	database.DB.Delete(&ch)
	response.Success(c, gin.H{"message": "channel deleted"})
}

func GetNotificationHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	channelID := c.Query("channel_id")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	db := database.DB.Model(&model.NotificationHistory{})
	if channelID != "" {
		db = db.Where("channel_id = ?", channelID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if startTime != "" {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("created_at <= ?", endTime)
	}

	var total int64
	db.Count(&total)

	var history []model.NotificationHistory
	db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&history)

	response.Paginated(c, history, int(total), page, pageSize)
}

func TestNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("channelId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid channel id")
		return
	}

	var ch model.NotificationChannel
	if err := database.DB.First(&ch, uint(id)).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "channel not found")
		return
	}

	if !ch.Enabled {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "channel is disabled")
		return
	}

	title, body := notify.FormatTestMessage(ch.Name)
	notify.SendAlert(ch.ID, "test", "info", title, body)

	response.Success(c, gin.H{"message": "test notification sent", "channel": ch.Name})
}
