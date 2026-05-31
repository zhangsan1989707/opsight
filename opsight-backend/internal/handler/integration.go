package handler

import (
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetIntegrations(c *gin.Context) {
	var integrations []model.Integration
	database.DB.Find(&integrations)

	result := make([]dto.IntegrationDTO, len(integrations))
	for i, item := range integrations {
		result[i] = dto.FromIntegration(item)
	}
	response.Success(c, gin.H{"integrations": result, "total": len(result)})
}
