package handler

import (
	"net/http"

	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetServices(c *gin.Context) {
	var services []model.Service
	database.DB.Find(&services)

	result := make([]dto.ServiceDTO, len(services))
	for i, s := range services {
		result[i] = dto.FromService(s)
	}
	response.Success(c, gin.H{"services": result, "total": len(result)})
}

func GetService(c *gin.Context) {
	name := c.Param("name")
	var s model.Service
	if err := database.DB.Where("name = ?", name).First(&s).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "service not found")
		return
	}
	response.Success(c, dto.FromService(s))
}

func CreateService(c *gin.Context) {
	var s model.Service
	if err := c.ShouldBindJSON(&s); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if s.Name == "" {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "service name is required")
		return
	}
	if s.Status == "" {
		s.Status = "healthy"
	}
	if err := database.DB.Create(&s).Error; err != nil {
		response.Error(c, http.StatusConflict, response.ErrBadRequest, "service already exists or invalid data")
		return
	}
	response.Success(c, dto.FromService(s))
}

func UpdateService(c *gin.Context) {
	name := c.Param("name")
	var s model.Service
	if err := database.DB.Where("name = ?", name).First(&s).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "service not found")
		return
	}

	var update model.Service
	if err := c.ShouldBindJSON(&update); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}

	database.DB.Model(&s).Updates(model.Service{
		Status:  update.Status,
		RPS:     update.RPS,
		P50:     update.P50,
		P99:     update.P99,
		ErrRate: update.ErrRate,
		Uptime:  update.Uptime,
		Team:    update.Team,
	})

	database.DB.Where("name = ?", name).First(&s)
	response.Success(c, dto.FromService(s))
}

func DeleteService(c *gin.Context) {
	name := c.Param("name")
	var s model.Service
	if err := database.DB.Where("name = ?", name).First(&s).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "service not found")
		return
	}

	database.DB.Where("service_name = ?", name).Delete(&model.ServiceDependency{})
	database.DB.Delete(&s)
	response.Success(c, gin.H{"message": "service deleted"})
}
