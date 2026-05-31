package handler

import (
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetInsights(c *gin.Context) {
	insightType := c.DefaultQuery("type", "root-cause")

	var items []model.Insight
	database.DB.Where("type = ?", insightType).Find(&items)

	result := make([]dto.InsightDTO, len(items))
	for i, item := range items {
		result[i] = dto.FromInsight(item)
	}
	response.Success(c, gin.H{"type": insightType, "insights": result})
}
