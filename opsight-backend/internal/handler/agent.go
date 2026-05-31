package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"opsight-backend/internal/database"
	"opsight-backend/internal/metrics"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// AgentAPIKeyAuth returns middleware that validates the Agent API key.
// If AGENT_API_KEY is not configured, it returns 401 for ALL requests (no bypass).
func AgentAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := os.Getenv("AGENT_API_KEY")
		if key == "" {
			logger.Error().Msg("AGENT_API_KEY is not configured - rejecting all agent reports")
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "agent authentication not configured")
			c.Abort()
			return
		}
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != key {
			logger.Warn().Str("client_ip", c.ClientIP()).Msg("invalid agent api key attempted")
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid agent api key")
			c.Abort()
			return
		}
		c.Next()
	}
}

type agentReportRequest struct {
	AgentVersion string `json:"agent_version"`
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	OS           string `json:"os"`
	CPU          struct {
		Cores   int     `json:"cores"`
		Percent float64 `json:"percent"`
	} `json:"cpu"`
	Memory struct {
		TotalMB float64 `json:"total_mb"`
		UsedMB  float64 `json:"used_mb"`
		Percent float64 `json:"percent"`
	} `json:"memory"`
	Disk struct {
		TotalMB float64 `json:"total_mb"`
		UsedMB  float64 `json:"used_mb"`
		Percent float64 `json:"percent"`
	} `json:"disk"`
	Network struct {
		RecvBytesPerSec float64 `json:"recv_bytes_per_sec"`
		SentBytesPerSec float64 `json:"sent_bytes_per_sec"`
	} `json:"network"`
	Load struct {
		Load1  float64 `json:"load1"`
		Load5  float64 `json:"load5"`
		Load15 float64 `json:"load15"`
	} `json:"load"`
}

func AgentReport(c *gin.Context) {
	var req agentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if req.Hostname == "" {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "hostname is required")
		return
	}
	if len(req.Hostname) > 255 {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "hostname too long")
		return
	}

	req.CPU.Percent = clampPercent(req.CPU.Percent)
	req.Memory.Percent = clampPercent(req.Memory.Percent)
	req.Disk.Percent = clampPercent(req.Disk.Percent)

	if req.Memory.TotalMB < 0 || req.Memory.UsedMB < 0 || req.Memory.UsedMB > req.Memory.TotalMB {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid memory values")
		return
	}
	if req.Disk.TotalMB < 0 || req.Disk.UsedMB < 0 || req.Disk.UsedMB > req.Disk.TotalMB {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid disk values")
		return
	}
	if req.Network.RecvBytesPerSec < 0 || req.Network.SentBytesPerSec < 0 {
		req.Network.RecvBytesPerSec = 0
		req.Network.SentBytesPerSec = 0
	}

	now := time.Now()

	// Upsert AgentInstance
	var agent model.AgentInstance
	result := database.DB.Where("hostname = ?", req.Hostname).First(&agent)
	if result.Error != nil {
		agent = model.AgentInstance{
			AgentVersion: req.AgentVersion,
			Hostname:     req.Hostname,
			IP:           req.IP,
			OS:           req.OS,
			CPUCores:     req.CPU.Cores,
			MemTotalMB:   req.Memory.TotalMB,
			Status:       "online",
			FirstSeenAt:  now,
			LastSeenAt:   now,
		}
		database.DB.Create(&agent)
	} else {
		agent.AgentVersion = req.AgentVersion
		agent.IP = req.IP
		agent.OS = req.OS
		agent.CPUCores = req.CPU.Cores
		agent.MemTotalMB = req.Memory.TotalMB
		agent.Status = "online"
		agent.LastSeenAt = now
		database.DB.Save(&agent)
	}

	// Insert MetricSnapshot
	snapshot := model.MetricSnapshot{
		AgentID:      agent.ID,
		Hostname:     req.Hostname,
		CPUPercent:   req.CPU.Percent,
		MemTotalMB:   req.Memory.TotalMB,
		MemUsedMB:    req.Memory.UsedMB,
		MemPercent:   req.Memory.Percent,
		DiskTotalMB:  req.Disk.TotalMB,
		DiskUsedMB:   req.Disk.UsedMB,
		DiskPercent:  req.Disk.Percent,
		NetRecvBytes: req.Network.RecvBytesPerSec,
		NetSentBytes: req.Network.SentBytesPerSec,
		Load1:        req.Load.Load1,
		Load5:        req.Load.Load5,
		Load15:       req.Load.Load15,
		ReportedAt:   now,
	}
	database.DB.Create(&snapshot)

	var totalAgents int64
	database.DB.Model(&model.AgentInstance{}).Count(&totalAgents)
	var onlineAgents int64
	database.DB.Model(&model.AgentInstance{}).Where("status = ?", "online").Count(&onlineAgents)
	metrics.SetAgentsTotal(int(totalAgents))
	metrics.SetAgentsOnline(int(onlineAgents))

	response.Success(c, gin.H{"message": "success"})
}

func ListAgents(c *gin.Context) {
	var agents []model.AgentInstance
	database.DB.Find(&agents)

	type agentSummary struct {
		ID         uint      `json:"id"`
		Hostname   string    `json:"hostname"`
		IP         string    `json:"ip"`
		OS         string    `json:"os"`
		Status     string    `json:"status"`
		LastSeenAt time.Time `json:"last_seen_at"`
	}
	result := make([]agentSummary, len(agents))
	for i, a := range agents {
		result[i] = agentSummary{
			ID:         a.ID,
			Hostname:   a.Hostname,
			IP:         a.IP,
			OS:         a.OS,
			Status:     a.Status,
			LastSeenAt: a.LastSeenAt,
		}
	}
	response.Success(c, gin.H{"agents": result, "total": len(result)})
}

func GetAgent(c *gin.Context) {
	hostname := c.Param("hostname")
	var agent model.AgentInstance
	if err := database.DB.Where("hostname = ?", hostname).First(&agent).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "agent not found")
		return
	}

	var latest model.MetricSnapshot
	database.DB.Where("agent_id = ?", agent.ID).Order("reported_at DESC").First(&latest)

	response.Success(c, gin.H{
		"agent":         agent,
		"latest_metric": latest,
	})
}

func GetAgentMetrics(c *gin.Context) {
	hostname := c.Param("hostname")
	metricType := c.DefaultQuery("metric", "cpu")
	duration := c.DefaultQuery("duration", "1h")
	limitStr := c.DefaultQuery("limit", "60")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 500 {
		limit = 60
	}

	var since time.Time
	switch duration {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "6h":
		since = time.Now().Add(-6 * time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	default:
		since = time.Now().Add(-1 * time.Hour)
	}

	var snapshots []model.MetricSnapshot
	database.DB.Where("hostname = ? AND reported_at >= ?", hostname, since).
		Order("reported_at ASC").Limit(limit).Find(&snapshots)

	type metricPoint struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
	}
	points := make([]metricPoint, len(snapshots))
	for i, s := range snapshots {
		var val float64
		switch metricType {
		case "cpu":
			val = s.CPUPercent
		case "memory":
			val = s.MemPercent
		case "disk":
			val = s.DiskPercent
		case "network_recv":
			val = s.NetRecvBytes
		case "network_sent":
			val = s.NetSentBytes
		case "load":
			val = s.Load1
		default:
			val = s.CPUPercent
		}
		points[i] = metricPoint{
			Timestamp: s.ReportedAt.Format(time.RFC3339),
			Value:     val,
		}
	}

	response.Success(c, gin.H{
		"hostname": hostname,
		"metric":   metricType,
		"points":   points,
	})
}
