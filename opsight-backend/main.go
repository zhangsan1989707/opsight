package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/internal/notify"
	"opsight-backend/pkg/logger"
	"opsight-backend/pkg/response"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ==================== DTOs (API response structs) ====================

type ServiceDTO struct {
	Name   string   `json:"name"`
	Status string   `json:"status"`
	RPS    string   `json:"rps"`
	P50    string   `json:"p50"`
	P99    string   `json:"p99"`
	ErrRate string  `json:"err_rate"`
	Uptime string   `json:"uptime"`
	Team   string   `json:"team"`
	Deps   []string `json:"deps"`
}

type IncidentDTO struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Service string `json:"service"`
	Status  string `json:"status"`
	Duration string `json:"duration"`
	Time    string `json:"time"`
	Detail  string `json:"detail,omitempty"`
}

type AlertRuleDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Threshold string `json:"threshold"`
	Service   string `json:"service"`
	Severity  string `json:"severity"`
	LastTrig  string `json:"last_triggered"`
	Enabled   bool   `json:"enabled"`
	IsAI      bool   `json:"is_ai"`
}

type InsightDTO struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Service    string `json:"service"`
	Confidence string `json:"confidence"`
	Time       string `json:"time"`
	Severity   string `json:"severity"`
	Related    string `json:"related,omitempty"`
}

type TopologyNodeDTO struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	RPS    string   `json:"rps"`
	P99    string   `json:"p99"`
	Deps   []string `json:"deps"`
}

type IntegrationDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	Enabled    bool   `json:"enabled"`
	EventCount int    `json:"event_count"`
}

type TeamMemberDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Team  string `json:"team"`
}

type TopErrorDTO struct {
	Error   string `json:"error"`
	Count   int    `json:"count"`
	Trend   string `json:"trend"`
	Service string `json:"service"`
}

// ==================== Agent API Key Middleware ====================

func agentAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := os.Getenv("AGENT_API_KEY")
		if key == "" {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != key {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid agent api key")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ==================== Agent Handlers ====================

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

func agentReport(c *gin.Context) {
	var req agentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if req.Hostname == "" {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "hostname is required")
		return
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

	logger.Info().
		Str("hostname", req.Hostname).
		Float64("cpu", req.CPU.Percent).
		Float64("mem", req.Memory.Percent).
		Msg("Agent report received")

	response.Success(c, gin.H{"message": "success"})
}

func listAgents(c *gin.Context) {
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

func getAgent(c *gin.Context) {
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

func getAgentMetrics(c *gin.Context) {
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
	db := database.DB.Where("hostname = ? AND reported_at >= ?", hostname, since).
		Order("reported_at ASC").Limit(limit)
	db.Find(&snapshots)

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

// ==================== WebSocket ====================

var (
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients    = make(map[*websocket.Conn]bool)
	clientsMu  sync.RWMutex
)

// ==================== Helpers ====================

var authSvc = auth.NewAuthService()

func serviceDeps(svcName string) []string {
	var deps []model.ServiceDependency
	database.DB.Where("service_name = ?", svcName).Find(&deps)
	result := make([]string, len(deps))
	for i, d := range deps {
		result[i] = d.DependencyID
	}
	return result
}

func topologyDeps(nodeName string) []string {
	var deps []model.TopologyDependency
	database.DB.Where("node_name = ?", nodeName).Find(&deps)
	result := make([]string, len(deps))
	for i, d := range deps {
		result[i] = d.DepNodeID
	}
	return result
}

func dbServiceToDTO(s model.Service) ServiceDTO {
	return ServiceDTO{
		Name:    s.Name,
		Status:  s.Status,
		RPS:     s.RPS,
		P50:     s.P50,
		P99:     s.P99,
		ErrRate: s.ErrRate,
		Uptime:  s.Uptime,
		Team:    s.Team,
		Deps:    serviceDeps(s.Name),
	}
}

func dbIncidentToDTO(i model.Incident) IncidentDTO {
	return IncidentDTO{
		ID:       i.ID,
		Summary:  i.Summary,
		Service:  i.Service,
		Status:   i.Status,
		Duration: i.Duration,
		Time:     i.Time,
		Detail:   i.Detail,
	}
}

func dbAlertRuleToDTO(r model.AlertRule) AlertRuleDTO {
	return AlertRuleDTO{
		ID:        r.ID,
		Name:      r.Name,
		Condition: r.Condition,
		Threshold: r.Threshold,
		Service:   r.Service,
		Severity:  r.Severity,
		LastTrig:  r.LastTriggered,
		Enabled:   r.Enabled,
		IsAI:      r.IsAI,
	}
}

func dbInsightToDTO(i model.Insight) InsightDTO {
	return InsightDTO{
		Type:       i.Type,
		Title:      i.Title,
		Body:       i.Body,
		Service:    i.Service,
		Confidence: i.Confidence,
		Time:       i.Time,
		Severity:   i.Severity,
		Related:    i.Related,
	}
}

func dbTopologyNodeToDTO(n model.TopologyNode) TopologyNodeDTO {
	return TopologyNodeDTO{
		ID:     n.Name,
		Status: n.Status,
		RPS:    n.RPS,
		P99:    n.P99,
		Deps:   topologyDeps(n.Name),
	}
}

func dbIntegrationToDTO(i model.Integration) IntegrationDTO {
	return IntegrationDTO{
		ID:         i.ID,
		Name:       i.Name,
		Type:       i.Type,
		Category:   i.Category,
		Status:     i.Status,
		Enabled:    i.Enabled,
		EventCount: i.EventCount,
	}
}

func dbTeamMemberToDTO(m model.TeamMember) TeamMemberDTO {
	return TeamMemberDTO{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
		Role:  m.Role,
		Team:  m.Team,
	}
}

func dbTopErrorToDTO(e model.TopError) TopErrorDTO {
	return TopErrorDTO{
		Error:   e.Error,
		Count:   e.Count,
		Trend:   e.Trend,
		Service: e.Service,
	}
}

// ==================== Handlers ====================

func healthCheck(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok", "time": time.Now().UTC()})
}

func getDashboardSummary(c *gin.Context) {
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

	response.Success(c, gin.H{
		"active_incidents":  activeIncidents,
		"mttr_minutes":      4.2,
		"services_healthy":  healthy,
		"services_degraded": degraded,
		"services_down":     down,
		"services_total":    len(services),
		"ai_alerts_today":   47,
		"ai_auto_resolved":  34,
	})
}

func getErrorRate(c *gin.Context) {
	labels := make([]string, 24)
	values := make([]float64, 24)
	data := []float64{0.12, 0.10, 0.08, 0.09, 0.11, 0.15, 0.18, 0.22, 0.31, 0.45, 0.38, 0.29, 0.24, 0.20, 0.18, 0.15, 0.21, 0.33, 0.41, 0.52, 0.38, 0.28, 0.19, 0.14}
	for i := 0; i < 24; i++ {
		labels[i] = fmt.Sprintf("%02d:00", i)
		values[i] = data[i]
	}
	response.Success(c, gin.H{"labels": labels, "values": values})
}

func getLatency(c *gin.Context) {
	svcLabels := []string{"api-gw", "auth", "payment", "user", "order", "notify", "cache", "events", "search", "cdn", "analytics", "mesh"}
	p50 := []int{12, 8500, 1200, 8, 18, 6, 85, 4, 22, 2, 35, 1}
	p90 := []int{28, 10000, 1800, 22, 45, 18, 150, 10, 55, 5, 80, 2}
	p99 := []int{45, 12000, 2100, 32, 67, 28, 210, 15, 89, 8, 120, 2}
	response.Success(c, gin.H{"labels": svcLabels, "p50": p50, "p90": p90, "p99": p99})
}

func getTopErrors(c *gin.Context) {
	var topErrors []model.TopError
	database.DB.Find(&topErrors)

	dto := make([]TopErrorDTO, len(topErrors))
	for i, e := range topErrors {
		dto[i] = dbTopErrorToDTO(e)
	}
	response.Success(c, gin.H{"errors": dto})
}

func getIncidents(c *gin.Context) {
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

	dto := make([]IncidentDTO, len(incidents))
	for i, inc := range incidents {
		dto[i] = dbIncidentToDTO(inc)
	}
	response.Success(c, gin.H{"incidents": dto, "total": len(dto)})
}

func getIncident(c *gin.Context) {
	id := c.Param("id")
	var inc model.Incident
	if err := database.DB.Where("id = ?", id).First(&inc).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "incident not found")
		return
	}
	response.Success(c, dbIncidentToDTO(inc))
}

func resolveIncident(c *gin.Context) {
	id := c.Param("id")
	var inc model.Incident
	if err := database.DB.Where("id = ?", id).First(&inc).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "incident not found")
		return
	}
	inc.Status = "resolved"
	inc.Duration = "resolved"
	database.DB.Save(&inc)

	// Audit log
	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "resolve", "incidents", id, "Incident resolved", c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dbIncidentToDTO(inc))
}

func getServices(c *gin.Context) {
	var services []model.Service
	database.DB.Find(&services)

	dto := make([]ServiceDTO, len(services))
	for i, s := range services {
		dto[i] = dbServiceToDTO(s)
	}
	response.Success(c, gin.H{"services": dto, "total": len(dto)})
}

func getService(c *gin.Context) {
	name := c.Param("name")
	var s model.Service
	if err := database.DB.Where("name = ?", name).First(&s).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "service not found")
		return
	}
	response.Success(c, dbServiceToDTO(s))
}

func getAlertRules(c *gin.Context) {
	var rules []model.AlertRule
	database.DB.Find(&rules)

	dto := make([]AlertRuleDTO, len(rules))
	for i, r := range rules {
		dto[i] = dbAlertRuleToDTO(r)
	}
	response.Success(c, gin.H{"rules": dto, "total": len(dto)})
}

func toggleAlertRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.AlertRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "rule not found")
		return
	}
	rule.Enabled = !rule.Enabled
	database.DB.Save(&rule)

	// Audit log
	userID, email, _ := auth.GetCurrentUser(c)
	action := "enabled"
	if !rule.Enabled {
		action = "disabled"
	}
	audit.Log(userID, email, "toggle", "alert-rules", id, "Alert rule "+action+": "+rule.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dbAlertRuleToDTO(rule))
}

func getMetricsQuery(c *gin.Context) {
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
			val = bucket[len(bucket)-1] // latest in bucket
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

func getMetricsNames(c *gin.Context) {
	names := []string{
		"cpu_usage", "memory_usage", "disk_usage",
		"network_recv", "network_sent", "load_avg",
		"error_rate", "latency_p50", "latency_p99",
		"request_rate", "connection_count", "gc_pause", "thread_count",
	}
	response.Success(c, gin.H{"metrics": names})
}

func getTopology(c *gin.Context) {
	var nodes []model.TopologyNode
	database.DB.Find(&nodes)

	dto := make([]TopologyNodeDTO, len(nodes))
	for i, n := range nodes {
		dto[i] = dbTopologyNodeToDTO(n)
	}
	sort.Slice(dto, func(i, j int) bool { return dto[i].ID < dto[j].ID })
	response.Success(c, gin.H{"nodes": dto})
}

func getRCA(c *gin.Context) {
	serviceID := c.Param("serviceId")

	type rcaResult struct {
		Service    string   `json:"service"`
		RootCause  string   `json:"root_cause"`
		Chain      []string `json:"chain"`
		Confidence string   `json:"confidence"`
	}

	var node model.TopologyNode
	if err := database.DB.Where("name = ?", serviceID).First(&node).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "service not found")
		return
	}

	if node.Status == "healthy" {
		response.Success(c, gin.H{"service": serviceID, "status": "healthy", "message": "No issues detected"})
		return
	}

	chain := []string{serviceID}
	rootCause := serviceID

	deps := topologyDeps(serviceID)
	for _, dep := range deps {
		var depNode model.TopologyNode
		if err := database.DB.Where("name = ?", dep).First(&depNode).Error; err == nil && depNode.Status != "healthy" {
			chain = append(chain, dep)
			rootCause = dep
		}
	}

	response.Success(c, rcaResult{
		Service:    serviceID,
		RootCause:  rootCause,
		Chain:      chain,
		Confidence: "94%",
	})
}

func getInsights(c *gin.Context) {
	insightType := c.DefaultQuery("type", "root-cause")

	var items []model.Insight
	database.DB.Where("type = ?", insightType).Find(&items)

	dto := make([]InsightDTO, len(items))
	for i, item := range items {
		dto[i] = dbInsightToDTO(item)
	}
	response.Success(c, gin.H{"type": insightType, "insights": dto})
}

func getIntegrations(c *gin.Context) {
	var integrations []model.Integration
	database.DB.Find(&integrations)

	dto := make([]IntegrationDTO, len(integrations))
	for i, item := range integrations {
		dto[i] = dbIntegrationToDTO(item)
	}
	response.Success(c, gin.H{"integrations": dto, "total": len(dto)})
}

func getTeam(c *gin.Context) {
	var members []model.TeamMember
	database.DB.Find(&members)

	dto := make([]TeamMemberDTO, len(members))
	for i, m := range members {
		dto[i] = dbTeamMemberToDTO(m)
	}
	response.Success(c, gin.H{"members": dto, "total": len(dto)})
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t2, err2 := time.Parse("2006-01-02", s)
		if err2 != nil {
			return time.Time{}
		}
		return t2
	}
	return t
}

func getAuditLogs(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID := uint(0)
	if userIDStr != "" {
		if v, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			userID = uint(v)
		}
	}

	action := c.Query("action")
	resource := c.Query("resource")
	startTime := parseTime(c.Query("start_time"))
	endTime := parseTime(c.Query("end_time"))

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	logs, total, err := audit.Query(userID, action, resource, startTime, endTime, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to query audit logs")
		return
	}

	response.Paginated(c, logs, int(total), page, pageSize)
}

func getAuditStats(c *gin.Context) {
	totalCount, todayCount, breakdown, err := audit.Stats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to get audit stats")
		return
	}

	actionsMap := make(map[string]int64)
	for _, b := range breakdown {
		actionsMap[b.Action] = b.Count
	}

	response.Success(c, gin.H{
		"total_count":       totalCount,
		"today_count":       todayCount,
		"actions_breakdown": actionsMap,
	})
}

func login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "email and password required")
		return
	}

	result, err := authSvc.Login(req.Email, req.Password)
	if err != nil {
		audit.Log(0, req.Email, "login_failed", "auth/login", "", err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), "failure")
		response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, err.Error())
		return
	}

	audit.Log(0, req.Email, "login", "auth/login", "", "Login successful", c.ClientIP(), c.GetHeader("User-Agent"), "success")
	response.Success(c, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

func register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "name, email, and password are required")
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}

	user, err := authSvc.Register(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		response.Error(c, http.StatusConflict, response.ErrBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"user": user})
}

func getCurrentUser(c *gin.Context) {
	userID, _, _ := auth.GetCurrentUser(c)
	user, err := authSvc.GetUserByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}
	response.Success(c, gin.H{"user": user})
}

func refreshToken(c *gin.Context) {
	userID, email, role := auth.GetCurrentUser(c)
	token, err := auth.GenerateToken(userID, email, role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to refresh token")
		return
	}
	response.Success(c, gin.H{"token": token})
}

// ==================== Notification Handlers ====================

func listNotificationChannels(c *gin.Context) {
	var channels []model.NotificationChannel
	database.DB.Find(&channels)
	response.Success(c, gin.H{"channels": channels})
}

func createNotificationChannel(c *gin.Context) {
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

func updateNotificationChannel(c *gin.Context) {
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

func deleteNotificationChannel(c *gin.Context) {
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

func getNotificationHistory(c *gin.Context) {
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

func testNotification(c *gin.Context) {
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

// ==================== WebSocket ====================

func handleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error().Err(err).Msg("WebSocket upgrade error")
		return
	}
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func broadcastEvent(eventType string, data interface{}) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	msg := gin.H{"type": eventType, "data": data, "time": time.Now().UTC()}
	for conn := range clients {
		conn.WriteJSON(msg)
	}
}

func evaluateAlerts() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		updateServiceMetrics()

		// 1. Query all enabled alert rules
		var rules []model.AlertRule
		database.DB.Where("enabled = ?", true).Find(&rules)
		if len(rules) == 0 {
			continue
		}

		// 2. Get all online agents with recent metrics
		var agents []model.AgentInstance
		database.DB.Where("status = ?", "online").Find(&agents)

		for _, rule := range rules {
			condition := strings.TrimSpace(rule.Condition)
			parts := strings.SplitN(condition, " > ", 2)
			if len(parts) != 2 {
				logger.Debug().Str("rule", rule.ID).Str("condition", condition).Msg("Rule condition unable to parse, skipping")
				continue
			}
			metricName := strings.TrimSpace(parts[0])
			thresholdStr := strings.TrimSpace(parts[1])

			// Extract numeric threshold
			threshold, err := strconv.ParseFloat(strings.TrimSuffix(thresholdStr, "%"), 64)
			if err != nil {
				logger.Debug().Str("rule", rule.ID).Str("threshold", thresholdStr).Msg("Rule threshold not numeric, skipping")
				continue
			}

			// For percentage thresholds (e.g. 85%), use as-is
			if strings.HasSuffix(thresholdStr, "%") {
				// threshold is already the numeric value (e.g. 85)
			}

			for _, agent := range agents {
				// Get latest metric for this agent
				var latest model.MetricSnapshot
				if err := database.DB.Where("agent_id = ?", agent.ID).Order("reported_at DESC").First(&latest).Error; err != nil {
					continue
				}

				// Skip if latest metric is too old (> 2 minutes)
				if time.Since(latest.ReportedAt) > 2*time.Minute {
					continue
				}

				var currentValue float64
				evaluatable := true

				switch {
				case strings.HasPrefix(metricName, "cpu_usage"):
					currentValue = latest.CPUPercent
				case strings.HasPrefix(metricName, "memory_usage"):
					currentValue = latest.MemPercent
				case strings.HasPrefix(metricName, "disk_usage"):
					currentValue = latest.DiskPercent
				case strings.HasPrefix(metricName, "p99"):
					currentValue = latest.Load1
				default:
					logger.Debug().Str("rule", rule.ID).Str("metric", metricName).Msg("Rule not yet evaluatable (service-level metric)")
					evaluatable = false
				}

				if !evaluatable {
					continue
				}

				// Check if already firing
				var existingAlert model.AlertEvent
				alreadyFiring := database.DB.Where(
					"alert_rule_id = ? AND hostname = ? AND status = ?",
					rule.ID, agent.Hostname, "firing",
				).First(&existingAlert).Error == nil

				if currentValue > threshold {
					if !alreadyFiring {
						// New alert — create AlertEvent
						msg := fmt.Sprintf("%s: %.1f > %.1f", rule.Name, currentValue, threshold)
						alertEvent := model.AlertEvent{
							AlertRuleID: rule.ID,
							RuleName:    rule.Name,
							Hostname:    agent.Hostname,
							Severity:    rule.Severity,
							Message:     msg,
							MetricValue: currentValue,
							Threshold:   threshold,
							Status:      "firing",
						}
						database.DB.Create(&alertEvent)

						// Update rule last_triggered
						rule.LastTriggered = "just now"
						database.DB.Model(&rule).Update("last_triggered", rule.LastTriggered)

						// Send notification
						sendNotificationForRule(rule, currentValue, threshold, agent.Hostname)

						logger.Info().
							Str("rule", rule.ID).
							Str("rule_name", rule.Name).
							Str("hostname", agent.Hostname).
							Float64("value", currentValue).
							Float64("threshold", threshold).
							Msg("Alert fired")

						// Broadcast to WebSocket
						broadcastEvent("alert_firing", gin.H{
							"rule_id":  rule.ID,
							"name":     rule.Name,
							"hostname": agent.Hostname,
							"severity": rule.Severity,
							"value":    currentValue,
							"threshold": threshold,
						})
					}
				} else {
					if alreadyFiring {
						// Resolve
						now := time.Now()
						database.DB.Model(&existingAlert).Updates(map[string]interface{}{
							"status":      "resolved",
							"resolved_at": now,
						})
						logger.Info().
							Str("rule", rule.ID).
							Str("hostname", agent.Hostname).
							Msg("Alert resolved")
					}
				}
			}
		}

		// Broadcast service_status periodically
		broadcastEvent("service_status", gin.H{"message": "Status update"})
	}
}

func sendNotificationForRule(rule model.AlertRule, value, threshold float64, hostname string) {
	var channels []model.NotificationChannel
	database.DB.Where("enabled = ?", true).Find(&channels)
	if len(channels) == 0 {
		return
	}

	for _, ch := range channels {
		title, body := notify.FormatAlertMessage(
			rule.Name,
			rule.Service,
			rule.Severity,
			rule.Condition,
			fmt.Sprintf("%.1f", threshold),
			fmt.Sprintf("%.1f", value),
			time.Now().Format(time.RFC3339),
		)

		logger.Info().
			Str("channel", ch.Name).
			Str("rule", rule.Name).
			Str("hostname", hostname).
			Str("severity", rule.Severity).
			Msg("Alert firing - sending notification")

		notify.SendAlert(ch.ID, rule.Name, rule.Severity, title, body)
	}
}

// ==================== Alert Event Handlers ====================

func listAlertEvents(c *gin.Context) {
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

func getAlertEvent(c *gin.Context) {
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

func resolveAlertEvent(c *gin.Context) {
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

func updateServiceMetrics() {
	var svcs []model.Service
	database.DB.Find(&svcs)
	for i := range svcs {
		if svcs[i].Status == "healthy" {
			rps := rand.Intn(500) + 1000
			svcs[i].RPS = fmt.Sprintf("%d", rps)
			p50 := rand.Intn(20) + 5
			svcs[i].P50 = fmt.Sprintf("%dms", p50)
			p99 := p50*3 + rand.Intn(50)
			svcs[i].P99 = fmt.Sprintf("%dms", p99)
			database.DB.Save(&svcs[i])
		}
	}

	var nodes []model.TopologyNode
	database.DB.Find(&nodes)
	for i := range nodes {
		if nodes[i].Status == "healthy" {
			rps := rand.Intn(500) + 1000
			nodes[i].RPS = fmt.Sprintf("%d", rps)
			p99 := rand.Intn(100) + 10
			nodes[i].P99 = fmt.Sprintf("%dms", p99)
			database.DB.Save(&nodes[i])
		}
	}
}

// ==================== Main ====================

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8800"
	}

	// Init database
	db := database.InitDB()
	// Seed data if tables are empty
	database.SeedAll(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.Middleware())
	r.Use(logger.GinLogger())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health
	r.GET("/healthz", healthCheck)

	// WebSocket
	r.GET("/api/v1/ws", handleWS)

	// API v1
	v1 := r.Group("/api/v1")
	v1.Use(audit.AuditMiddleware())
	{
		// Auth (public)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", login)

			// Protected auth routes
			authProtected := authGroup.Group("")
			authProtected.Use(auth.AuthRequired())
			{
				authProtected.POST("/register", auth.RequireRole("admin"), register)
				authProtected.GET("/me", getCurrentUser)
				authProtected.POST("/refresh", refreshToken)
			}
		}

		// Agent report (API key auth, not JWT)
		v1.POST("/agents/report", agentAPIKeyAuth(), agentReport)

		// Protected routes
		protected := v1.Group("")
		protected.Use(auth.AuthRequired())
		{
			// Dashboard
			protected.GET("/dashboard/summary", getDashboardSummary)
			protected.GET("/dashboard/error-rate", getErrorRate)
			protected.GET("/dashboard/latency", getLatency)
			protected.GET("/dashboard/top-errors", getTopErrors)

			// Incidents
			protected.GET("/incidents", getIncidents)
			protected.GET("/incidents/:id", getIncident)
			protected.POST("/incidents/:id/resolve", auth.RequireRole("admin", "editor"), resolveIncident)

			// Services
			protected.GET("/services", getServices)
			protected.GET("/services/:name", getService)

			// Alert Rules
			protected.GET("/alert-rules", getAlertRules)
			protected.PATCH("/alert-rules/:id/toggle", auth.RequireRole("admin", "editor"), toggleAlertRule)

			// Metrics
			protected.GET("/metrics/query", getMetricsQuery)
			protected.GET("/metrics/names", getMetricsNames)

			// Agents
			protected.GET("/agents", auth.RequireRole("admin"), listAgents)
			protected.GET("/agents/:hostname", auth.RequireRole("admin"), getAgent)
			protected.GET("/agents/:hostname/metrics", auth.RequireRole("admin"), getAgentMetrics)

			// Alert Events
			protected.GET("/alerts/events", listAlertEvents)
			protected.GET("/alerts/events/:id", getAlertEvent)
			protected.POST("/alerts/events/:id/resolve", auth.RequireRole("admin", "editor"), resolveAlertEvent)

			// Topology
			protected.GET("/topology", getTopology)
			protected.GET("/topology/:serviceId/rca", getRCA)

			// Insights
			protected.GET("/insights", getInsights)

			// Integrations
			protected.GET("/integrations", getIntegrations)

			// Team
			protected.GET("/team", auth.RequireRole("admin"), getTeam)

			// Audit Logs
			protected.GET("/audit-logs", auth.RequireRole("admin"), getAuditLogs)
			protected.GET("/audit-logs/stats", getAuditStats)

			// Notifications
			protected.GET("/notifications/channels", listNotificationChannels)
			protected.POST("/notifications/channels", auth.RequireRole("admin"), createNotificationChannel)
			protected.PUT("/notifications/channels/:id", auth.RequireRole("admin"), updateNotificationChannel)
			protected.DELETE("/notifications/channels/:id", auth.RequireRole("admin"), deleteNotificationChannel)
			protected.GET("/notifications/history", getNotificationHistory)
			protected.POST("/notifications/test/:channelId", testNotification)
		}
	}

	// Start alert evaluator
	go evaluateAlerts()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	logger.Info().Str("port", port).Msg("Opsight API starting")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Server error")
		os.Exit(1)
	}
}
