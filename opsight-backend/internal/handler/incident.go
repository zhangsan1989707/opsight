package handler

import (
	"net/http"
	"strings"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetIncidents(c *gin.Context) {
	status := c.Query("status")
	service := c.Query("service")
	search := c.Query("search")

	db := database.DB.Model(&model.Incident{})
	if status != "" && status != "all" {
		db = db.Where("status = ?", status)
	}
	if service != "" && service != "all" {
		db = db.Where("service = ?", service)
	}
	if search != "" {
		db = db.Where("LOWER(summary) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	var incidents []model.Incident
	db.Find(&incidents)

	result := make([]dto.IncidentDTO, len(incidents))
	for i, inc := range incidents {
		result[i] = dto.FromIncident(inc)
	}
	response.Success(c, gin.H{"incidents": result, "total": len(result)})
}

func GetIncident(c *gin.Context) {
	id := c.Param("id")
	var inc model.Incident
	if err := database.DB.Where("id = ?", id).First(&inc).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "incident not found")
		return
	}
	response.Success(c, dto.FromIncident(inc))
}

func ResolveIncident(c *gin.Context) {
	id := c.Param("id")
	var inc model.Incident
	if err := database.DB.Where("id = ?", id).First(&inc).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "incident not found")
		return
	}
	inc.Status = "resolved"
	inc.Duration = "resolved"
	database.DB.Save(&inc)

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "resolve", "incidents", id, "Incident resolved", c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dto.FromIncident(inc))
}
