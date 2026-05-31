package handler

import (
	"net/http"
	"time"

	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok", "time": time.Now().UTC()})
}

func NotFound(c *gin.Context) {
	response.Error(c, http.StatusNotFound, response.ErrNotFound, "endpoint not found")
}
