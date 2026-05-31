package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"opsight-backend/internal/cache"
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetDashboardSummary(c *gin.Context) {
	cacheKey := "opsight:dashboard:summary"
	var cacheData gin.H

	if hit, _ := cache.GetJSON(c, cacheKey, &cacheData); hit {
		response.Success(c, cacheData)
		return
	}

	var services []model.Service
	database.DB.Find(&services)

	healthy, degraded, down := 0, 0, 0
	for _, s := range services {
		switch s.Status {
		case "healthy":
			healthy++
		case "degraded":
			degraded++
		case "down":
			down++
		}
	}

	var activeIncidents int64
	database.DB.Model(&model.Incident{}).Where("status IN ?", []string{"critical", "warning"}).Count(&activeIncidents)

	var resolvedIncidents int64
	database.DB.Model(&model.Incident{}).Where("status = ?", "resolved").Count(&resolvedIncidents)

	today := time.Now().Truncate(24 * time.Hour)
	var alertsToday int64
	database.DB.Model(&model.AlertEvent{}).Where("created_at >= ?", today).Count(&alertsToday)

	var autoResolved int64
	database.DB.Model(&model.AlertEvent{}).Where("created_at >= ? AND status = ?", today, "resolved").Count(&autoResolved)

	var avgMTTR float64
	type mttrRow struct {
		Duration float64
	}
	var result mttrRow
	database.DB.Model(&model.AlertEvent{}).
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60.0) as duration").
		Where("status = ? AND resolved_at IS NOT NULL AND created_at >= ?", "resolved", today).
		Scan(&result)
	avgMTTR = result.Duration
	if avgMTTR <= 0 {
		avgMTTR = 0
	}

	data := gin.H{
		"active_incidents":  activeIncidents,
		"mttr_minutes":      fmt.Sprintf("%.1f", avgMTTR),
		"resolved_total":    resolvedIncidents,
		"services_healthy":  healthy,
		"services_degraded": degraded,
		"services_down":     down,
		"services_total":    len(services),
		"ai_alerts_today":   alertsToday,
		"ai_auto_resolved":  autoResolved,
	}

	cache.SetJSON(c, cacheKey, data, 30*time.Second)
	response.Success(c, data)
}

func GetErrorRate(c *gin.Context) {
	now := time.Now()
	labels := make([]string, 24)
	values := make([]float64, 24)

	// Aggregate hourly avg CPU from metric snapshots (last 24h) as a proxy metric
	cutoff := now.Add(-24 * time.Hour)
	var snapshots []model.MetricSnapshot
	database.DB.Where("reported_at >= ?", cutoff).Order("reported_at ASC").Find(&snapshots)

	hourBuckets := make(map[int][]float64)
	for _, s := range snapshots {
		hourBuckets[s.ReportedAt.Hour()] = append(hourBuckets[s.ReportedAt.Hour()], s.CPUPercent)
	}

	for i := 0; i < 24; i++ {
		t := now.Add(time.Duration(-23+i) * time.Hour)
		labels[i] = t.Format("15:04")
		bucket := hourBuckets[t.Hour()]
		if len(bucket) > 0 {
			sum := 0.0
			for _, v := range bucket {
				sum += v
			}
			values[i] = math.Round(sum/float64(len(bucket))*100) / 100
		}
	}

	response.Success(c, gin.H{"labels": labels, "values": values})
}

func GetLatency(c *gin.Context) {
	var services []model.Service
	database.DB.Find(&services)

	labels := make([]string, len(services))
	p50 := make([]int, len(services))
	p99 := make([]int, len(services))

	for i, s := range services {
		labels[i] = s.Name
		p50[i] = parseMs(s.P50)
		p99[i] = parseMs(s.P99)
	}

	response.Success(c, gin.H{"labels": labels, "p50": p50, "p99": p99})
}

// parseMs extracts numeric value from strings like "12ms" or "4,200".
func parseMs(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "ms")
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.Atoi(s)
	return v
}

func GetTopErrors(c *gin.Context) {
	var topErrors []model.TopError
	database.DB.Find(&topErrors)

	result := make([]dto.TopErrorDTO, len(topErrors))
	for i, e := range topErrors {
		result[i] = dto.FromTopError(e)
	}
	response.Success(c, gin.H{"errors": result})
}
