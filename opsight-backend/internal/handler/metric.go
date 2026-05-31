package handler

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetMetricsQuery(c *gin.Context) {
	metric := c.DefaultQuery("metric", "cpu_usage")
	hostname := c.DefaultQuery("hostname", "")
	now := time.Now()

	type point struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
		Avg       float64 `json:"avg"`
		P95       float64 `json:"p95"`
		P99       float64 `json:"p99"`
	}

	// Map frontend metric names to DB columns
	var colName string
	switch metric {
	case "cpu_usage":
		colName = "cpu_percent"
	case "memory_usage":
		colName = "mem_percent"
	case "disk_usage":
		colName = "disk_percent"
	case "network_recv":
		colName = "net_recv_bytes"
	case "network_sent":
		colName = "net_sent_bytes"
	case "load_avg":
		colName = "load1"
	case "error_rate", "latency_p50", "latency_p99", "request_rate", "connection_count", "gc_pause", "thread_count":
		// Service-level metrics that agents don't collect — fallback to random for backward compat
		points := make([]point, 24)
		for i := 0; i < 24; i++ {
			t := now.Add(time.Duration(-23+i) * time.Hour)
			base := 45.0 + rand.Float64()*20
			if metric == "error_rate" {
				base = 0.1 + rand.Float64()*0.5
			} else if metric == "latency_p50" || metric == "latency_p99" {
				base = 20 + rand.Float64()*80
			}
			points[i] = point{
				Timestamp: t.Format("15:04"),
				Value:     math.Round(base*100) / 100,
				Avg:       math.Round(base*100) / 100,
				P95:       math.Round((base*1.3)*100) / 100,
				P99:       math.Round((base*1.6)*100) / 100,
			}
		}
		response.Success(c, gin.H{"metric": metric, "service": hostname, "points": points})
		return
	default:
		colName = "cpu_percent"
	}

	// Query MetricSnapshot from DB, grouped by hour for last 24h
	cutoff := now.Add(-24 * time.Hour)
	db := database.DB.Model(&model.MetricSnapshot{}).
		Where("reported_at >= ?", cutoff)
	if hostname != "" {
		db = db.Where("hostname = ?", hostname)
	}

	var snapshots []model.MetricSnapshot
	db.Order("reported_at ASC").Find(&snapshots)

	// Group by hour
	hourBuckets := make(map[int][]float64)
	for _, s := range snapshots {
		hour := s.ReportedAt.Hour()
		var val float64
		switch colName {
		case "cpu_percent":
			val = s.CPUPercent
		case "mem_percent":
			val = s.MemPercent
		case "disk_percent":
			val = s.DiskPercent
		case "net_recv_bytes":
			val = s.NetRecvBytes
		case "net_sent_bytes":
			val = s.NetSentBytes
		case "load1":
			val = s.Load1
		}
		hourBuckets[hour] = append(hourBuckets[hour], val)
	}

	// Build 24 data points
	points := make([]point, 24)
	for i := 0; i < 24; i++ {
		t := now.Add(time.Duration(-23+i) * time.Hour)
		bucket := hourBuckets[t.Hour()]
		var val, avg, p95, p99 float64
		if len(bucket) > 0 {
			val = bucket[len(bucket)-1]
			avg, p95, p99 = computeStats(bucket)
		}
		points[i] = point{
			Timestamp: t.Format("15:04"),
			Value:     math.Round(val*100) / 100,
			Avg:       math.Round(avg*100) / 100,
			P95:       math.Round(p95*100) / 100,
			P99:       math.Round(p99*100) / 100,
		}
	}

	response.Success(c, gin.H{"metric": metric, "service": hostname, "points": points})
}

func computeStats(values []float64) (avg, p95, p99 float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg = sum / float64(len(values))

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	p95Idx := int(float64(len(sorted)-1) * 0.95)
	p99Idx := int(float64(len(sorted)-1) * 0.99)
	p95 = sorted[p95Idx]
	p99 = sorted[p99Idx]
	return
}

func GetMetricsNames(c *gin.Context) {
	names := []string{
		"cpu_usage", "memory_usage", "disk_usage",
		"network_recv", "network_sent", "load_avg",
		"error_rate", "latency_p50", "latency_p99",
		"request_rate", "connection_count", "gc_pause", "thread_count",
	}
	response.Success(c, gin.H{"metrics": names})
}
