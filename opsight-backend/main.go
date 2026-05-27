package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ==================== Models ====================

type Service struct {
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	RPS      string   `json:"rps"`
	P50      string   `json:"p50"`
	P99      string   `json:"p99"`
	ErrRate  string   `json:"err_rate"`
	Uptime   string   `json:"uptime"`
	Team     string   `json:"team"`
	Deps     []string `json:"deps"`
}

type Incident struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Service  string `json:"service"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Time     string `json:"time"`
	Detail   string `json:"detail,omitempty"`
}

type AlertRule struct {
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

type Insight struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Service    string `json:"service"`
	Confidence string `json:"confidence"`
	Time       string `json:"time"`
	Severity   string `json:"severity"`
	Related    string `json:"related,omitempty"`
}

type TopologyNode struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	RPS    string   `json:"rps"`
	P99    string   `json:"p99"`
	Deps   []string `json:"deps"`
}

type Integration struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	Enabled    bool   `json:"enabled"`
	EventCount int    `json:"event_count"`
}

type TeamMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Team  string `json:"team"`
}

type TopError struct {
	Error   string `json:"error"`
	Count   int    `json:"count"`
	Trend   string `json:"trend"`
	Service string `json:"service"`
}

// ==================== Data Store ====================

var (
	mu         sync.RWMutex
	services   []Service
	incidents  []Incident
	alertRules []AlertRule
	insights   map[string][]Insight
	topology   map[string]TopologyNode
	integrations []Integration
	members    []TeamMember
	topErrors  []TopError

	// WebSocket
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[*websocket.Conn]bool)
)

func initData() {
	services = []Service{
		{"api-gateway", "healthy", "4,200", "12ms", "45ms", "0.02%", "99.99%", "Platform", []string{"user-service", "auth-service", "payment-gateway", "order-service", "cache-layer"}},
		{"auth-service", "down", "1,100", "8.5s", "12s", "14.2%", "99.81%", "Identity", []string{"cache-layer", "user-service"}},
		{"payment-gateway", "degraded", "890", "1.2s", "2.1s", "3.8%", "99.92%", "Payments", []string{"cache-layer", "event-pipeline"}},
		{"user-service", "healthy", "3,400", "8ms", "32ms", "0.01%", "99.98%", "Identity", []string{"cache-layer"}},
		{"order-service", "healthy", "2,100", "18ms", "67ms", "0.05%", "99.97%", "Commerce", []string{"payment-gateway", "event-pipeline", "user-service"}},
		{"notification-svc", "healthy", "560", "6ms", "28ms", "0.03%", "99.95%", "Platform", []string{"event-pipeline"}},
		{"cache-layer", "degraded", "12,000", "2ms", "210ms", "0.8%", "99.88%", "Infrastructure", []string{}},
		{"event-pipeline", "healthy", "8,700", "4ms", "15ms", "0.01%", "99.99%", "Data", []string{}},
		{"search-service", "healthy", "1,800", "22ms", "89ms", "0.08%", "99.94%", "Discovery", []string{"cache-layer"}},
		{"cdn-edge", "healthy", "45,000", "2ms", "8ms", "0.00%", "100.00%", "Infrastructure", []string{}},
		{"analytics-svc", "healthy", "2,300", "35ms", "120ms", "0.12%", "99.91%", "Data", []string{"event-pipeline"}},
		{"service-mesh", "healthy", "-", "0.5ms", "2ms", "0.00%", "99.99%", "Infrastructure", []string{}},
	}

	incidents = []Incident{
		{"INC-4221", "Memory leak in auth-svc causing OOM kills", "auth-service", "critical", "14m", "2 min ago", "auth-svc v2.4.1 disabled session-cache eviction. Memory grows 12 MB/min."},
		{"INC-4220", "Elevated 5xx on /api/v2/payments", "payment-gateway", "critical", "8m", "12 min ago", "Redis latency spike correlated with payment-gateway 5xx errors after deploy at 14:05 UTC."},
		{"INC-4219", "Redis cluster latency spike > 200ms", "cache-layer", "warning", "23m", "28 min ago", "Shared Redis instance showing disk IO saturation."},
		{"INC-4218", "Disk usage > 90% on us-east-1 node 7", "infra/storage", "warning", "1h 12m", "1 hr ago", "Log rotation disabled after infra update. Growth rate 2.1 GB/hr."},
		{"INC-4217", "Kafka consumer lag exceeds threshold", "event-pipeline", "resolved", "45m", "2 hr ago", "Consumer group scaled from 3 to 6. Lag resolved."},
		{"INC-4216", "DNS resolution failure for internal registry", "service-mesh", "resolved", "12m", "3 hr ago", "CoreDNS OOM killed after memory limit reduction."},
		{"INC-4215", "SSL certificate expiring on api-gateway", "api-gateway", "info", "-", "4 hr ago", "Certificate expires in 14 days."},
		{"INC-4214", "Pod restart loop in notification-svc", "notification-service", "resolved", "38m", "5 hr ago", "Missing env vars after config migration."},
	}

	alertRules = []AlertRule{
		{"AR-001", "High CPU Usage", "cpu_usage > 85", "85%", "All Services", "warning", "12 min ago", true, false},
		{"AR-002", "Memory Leak Detection", "memory_growth > 10MB/min", "10 MB/min", "auth-service", "critical", "2 min ago", true, true},
		{"AR-003", "Error Rate Spike", "error_rate > 5", "5%", "All Services", "critical", "12 min ago", true, false},
		{"AR-004", "Latency Anomaly", "p99 > 3x baseline", "3x", "All Services", "warning", "28 min ago", true, true},
		{"AR-005", "Disk Usage Warning", "disk_usage > 90", "90%", "infra/*", "warning", "1 hr ago", true, false},
		{"AR-006", "SSL Certificate Expiry", "cert_days < 14", "14 days", "All Services", "info", "4 hr ago", true, false},
		{"AR-007", "Pod Restart Loop", "restart_count > 5 in 10m", "5/10min", "All Services", "critical", "5 hr ago", true, true},
		{"AR-008", "Kafka Consumer Lag", "consumer_lag > 10000", "10,000", "event-pipeline", "warning", "2 hr ago", true, false},
		{"AR-009", "Connection Pool Exhaustion", "pool_usage > 90", "90%", "payment-gateway", "critical", "12 min ago", true, true},
		{"AR-010", "DNS Resolution Failure", "dns_fail_rate > 1", "1%", "service-mesh", "critical", "3 hr ago", false, false},
		{"AR-011", "Request Timeout Rate", "timeout_rate > 2", "2%", "api-gateway", "warning", "-", true, false},
		{"AR-012", "Cache Hit Ratio Drop", "cache_hit < 80", "80%", "cache-layer", "warning", "28 min ago", true, true},
		{"AR-013", "Deployment Anomaly", "post_deploy_error > 2x", "2x", "All Services", "critical", "12 min ago", true, true},
		{"AR-014", "Network Saturation", "bandwidth > 80", "80%", "cdn-edge", "info", "-", true, false},
		{"AR-015", "Service Discovery Failures", "sd_fail > 0", "0", "service-mesh", "warning", "3 hr ago", true, false},
		{"AR-016", "Queue Depth Alert", "queue_depth > 5000", "5,000", "event-pipeline", "warning", "2 hr ago", true, false},
		{"AR-017", "Response Size Anomaly", "response_kb > 10x avg", "10x", "api-gateway", "info", "-", false, true},
		{"AR-018", "GC Pause Time", "gc_pause > 200ms", "200ms", "auth-service", "warning", "2 min ago", true, false},
		{"AR-019", "Thread Pool Saturation", "thread_usage > 95", "95%", "payment-gateway", "critical", "12 min ago", true, true},
		{"AR-020", "Health Check Failures", "health_fail > 3 consecutive", "3", "All Services", "critical", "5 hr ago", true, false},
		{"AR-021", "API Rate Limit Approaching", "rate_pct > 80", "80%", "api-gateway", "info", "-", true, false},
		{"AR-022", "Database Connection Leak", "db_conn_growth > 10/hr", "10/hr", "user-service", "warning", "-", false, true},
		{"AR-023", "Certificate Chain Invalid", "cert_valid == false", "false", "api-gateway", "critical", "-", true, false},
		{"AR-024", "Memory Fragmentation", "mem_frag > 40", "40%", "cache-layer", "info", "28 min ago", true, true},
		{"AR-025", "Cross-Region Latency", "cross_region_p99 > 500ms", "500ms", "cdn-edge", "warning", "-", true, false},
		{"AR-026", "Secret Rotation Overdue", "secret_age > 90d", "90 days", "All Services", "info", "-", false, false},
		{"AR-027", "Retry Storm Detection", "retry_rate > 30", "30%", "order-service", "warning", "-", true, true},
		{"AR-028", "Log Volume Spike", "log_eps > 50000", "50K/s", "All Services", "info", "-", true, false},
		{"AR-029", "Dependency Health", "dep_fail > 0", "0", "api-gateway", "critical", "12 min ago", true, false},
		{"AR-030", "Anomalous Traffic Pattern", "traffic_zscore > 3", "3σ", "api-gateway", "warning", "-", true, true},
		{"AR-031", "Container Restart Storm", "restart_storm == true", "true", "All Services", "critical", "5 hr ago", true, true},
		{"AR-032", "Config Drift Detected", "config_hash != expected", "match", "All Services", "warning", "-", false, false},
		{"AR-033", "Cost Anomaly", "hourly_cost > 2x avg", "2x", "All Services", "info", "-", true, false},
		{"AR-034", "Data Freshness", "data_lag > 5min", "5 min", "analytics-svc", "warning", "-", true, false},
	}

	insights = map[string][]Insight{
		"root-cause": {
			{"critical", "Memory leak caused by session-cache eviction disabled in auth-svc v2.4.1", "Deploy at 14:05 UTC changed session.cache.eviction.enabled from true to false. Memory grows at 12 MB/min. Affected 3 pods across us-east-1.", "auth-service", "97%", "2 min ago", "critical", "INC-4221"},
			{"critical", "Redis latency spike correlated with payment-gateway 5xx errors", "Both anomalies triggered simultaneously after deploy at 14:05 UTC. Redis primary node showing disk IO saturation on the shared cache-layer instance.", "payment-gateway", "94%", "12 min ago", "critical", "INC-4220"},
			{"warning", "Kafka consumer lag from scaled-down group not restored", "Consumer group was manually scaled to 3 for maintenance window at 11:00 UTC. Auto-scaler was paused and not re-enabled. Lag grew from 0 to 12,000 over 90 minutes.", "event-pipeline", "99%", "2 hr ago", "warning", "INC-4217"},
			{"info", "DNS resolution failure from CoreDNS OOM kill", "CoreDNS pod was killed by OOM after memory limit reduced from 512Mi to 256Mi during cluster upgrade.", "service-mesh", "91%", "3 hr ago", "info", "INC-4216"},
		},
		"predictions": {
			{"warning", "us-east-1 node 7 disk will reach 95% in ~6 hours", "Current growth rate: 2.1 GB/hr. Log rotation disabled after last infra update.", "infra/storage", "88%", "1 hr ago", "warning", "INC-4218"},
			{"info", "auth-service failure may cascade to order-service within 30 min", "order-service depends on auth-service for token validation. Current auth-svc error rate of 14.2% will likely exhaust order-service circuit breaker.", "order-service", "76%", "5 min ago", "info", "INC-4221"},
			{"info", "SSL certificate for payment-vendor API expires in 14 days", "Certificate for payments-vendor.internal expires on 2026-06-10. Cert-manager has not initiated renewal.", "payment-gateway", "100%", "6 hr ago", "info", ""},
		},
		"remediation": {
			{"resolved", "Auto-scaled Kafka consumer group from 3 to 6 instances", "Consumer lag dropped from 12,000 to under 100 within 15 minutes.", "event-pipeline", "99%", "2 hr ago", "resolved", "INC-4217"},
			{"resolved", "Auto-restarted CoreDNS with restored memory limit", "Restored memory limit from 256Mi to 512Mi. DNS resolution returned to normal.", "service-mesh", "95%", "3 hr ago", "resolved", "INC-4216"},
			{"resolved", "Triggered log rotation and freed 12 GB on node 7", "Disk usage dropped from 91.2% to 78.9%.", "infra/storage", "90%", "1 hr ago", "resolved", "INC-4218"},
			{"info", "Suggested rollback for auth-svc to v2.4.0", "Rollback will re-enable session-cache eviction and resolve the memory leak.", "auth-service", "97%", "2 min ago", "info", "INC-4221"},
		},
		"patterns": {
			{"warning", "Shared Redis instance creates blast radius across services", "auth-service, payment-gateway, and user-service all share the same Redis cluster.", "cache-layer", "85%", "30 min ago", "warning", ""},
			{"info", "Deploys between 14:00-15:00 UTC have 3x higher incident rate", "67% of incidents coincide with deploys during this window.", "platform", "82%", "1 day ago", "info", ""},
			{"info", "notification-svc restarts correlate with config migrations", "3 of the last 4 config migrations triggered restart loops.", "notification-service", "78%", "5 hr ago", "info", "INC-4214"},
		},
	}

	topology = map[string]TopologyNode{
		"api-gateway":     {"api-gateway", "healthy", "4,200", "45ms", []string{"auth-service", "user-service", "payment-gateway", "order-service", "cache-layer"}},
		"auth-service":    {"auth-service", "down", "1,100", "12s", []string{"cache-layer", "user-service"}},
		"payment-gateway": {"payment-gateway", "degraded", "890", "2.1s", []string{"cache-layer", "event-pipeline"}},
		"cache-layer":     {"cache-layer", "degraded", "12,000", "210ms", []string{}},
		"user-service":    {"user-service", "healthy", "3,400", "32ms", []string{"cache-layer"}},
		"notification-svc":{"notification-svc", "healthy", "560", "28ms", []string{"event-pipeline"}},
		"order-service":   {"order-service", "healthy", "2,100", "67ms", []string{"payment-gateway", "event-pipeline", "user-service"}},
		"search-service":  {"search-service", "healthy", "1,800", "89ms", []string{"cache-layer"}},
		"analytics-svc":   {"analytics-svc", "healthy", "2,300", "120ms", []string{"event-pipeline"}},
		"event-pipeline":  {"event-pipeline", "healthy", "8,700", "15ms", []string{}},
		"cdn-edge":        {"cdn-edge", "healthy", "45,000", "8ms", []string{}},
		"service-mesh":    {"service-mesh", "healthy", "-", "2ms", []string{}},
	}

	integrations = []Integration{
		{"INT-001", "Slack Alerts", "slack", "Notification", "connected", true, 1247},
		{"INT-002", "PagerDuty", "pagerduty", "On-call", "connected", true, 892},
		{"INT-003", "Prometheus", "prometheus", "Metrics", "connected", true, 45200},
		{"INT-004", "OpenTelemetry", "otel", "Traces", "connected", true, 23100},
		{"INT-005", "Zabbix", "zabbix", "Legacy Monitoring", "connected", true, 634},
		{"INT-006", "Grafana", "grafana", "Visualization", "disconnected", false, 0},
		{"INT-007", "Webhook (Custom)", "webhook", "Custom", "connected", true, 201},
		{"INT-008", "Filebeat", "filebeat", "Logs", "connected", true, 89400},
	}

	members = []TeamMember{
		{"U001", "Leo Hang", "leo@opsight.io", "Admin", "Platform"},
		{"U002", "Zhang Wei", "zhangwei@opsight.io", "Editor", "Identity"},
		{"U003", "Li Na", "lina@opsight.io", "Editor", "Payments"},
		{"U004", "Wang Fang", "wangfang@opsight.io", "Viewer", "Commerce"},
		{"U005", "Chen Jie", "chenjie@opsight.io", "Editor", "Infrastructure"},
		{"U006", "Liu Yang", "liuyang@opsight.io", "Editor", "Data"},
		{"U007", "Zhao Min", "zhaomin@opsight.io", "Viewer", "Discovery"},
		{"U008", "Sun Lei", "sunlei@opsight.io", "Editor", "Platform"},
	}

	topErrors = []TopError{
		{"OutOfMemoryError: Java heap space", 1247, "up", "auth-service"},
		{"ConnectionPoolTimeoutException", 892, "up", "payment-gateway"},
		{"RedisTimeoutException: Command timed out", 634, "stable", "cache-layer"},
		{"KafkaException: OffsetOutOfRange", 201, "down", "event-pipeline"},
		{"SSLHandshakeException: Remote host closed", 89, "down", "api-gateway"},
	}
}

// ==================== Handlers ====================

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now().UTC()})
}

func getDashboardSummary(c *gin.Context) {
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
	activeIncidents := 0
	for _, inc := range incidents {
		if inc.Status == "critical" || inc.Status == "warning" {
			activeIncidents++
		}
	}
	c.JSON(200, gin.H{
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
	c.JSON(200, gin.H{"labels": labels, "values": values})
}

func getLatency(c *gin.Context) {
	svcLabels := []string{"api-gw", "auth", "payment", "user", "order", "notify", "cache", "events", "search", "cdn", "analytics", "mesh"}
	p50 := []int{12, 8500, 1200, 8, 18, 6, 85, 4, 22, 2, 35, 1}
	p90 := []int{28, 10000, 1800, 22, 45, 18, 150, 10, 55, 5, 80, 2}
	p99 := []int{45, 12000, 2100, 32, 67, 28, 210, 15, 89, 8, 120, 2}
	c.JSON(200, gin.H{"labels": svcLabels, "p50": p50, "p90": p90, "p99": p99})
}

func getTopErrors(c *gin.Context) {
	c.JSON(200, gin.H{"errors": topErrors})
}

func getIncidents(c *gin.Context) {
	status := c.Query("status")
	service := c.Query("service")
	search := c.Query("search")

	mu.RLock()
	defer mu.RUnlock()

	result := make([]Incident, 0, len(incidents))
	for _, inc := range incidents {
		if status != "" && status != "all" && inc.Status != status {
			continue
		}
		if service != "" && service != "all" && inc.Service != service {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(inc.Summary), strings.ToLower(search)) {
			continue
		}
		result = append(result, inc)
	}
	c.JSON(200, gin.H{"incidents": result, "total": len(result)})
}

func getIncident(c *gin.Context) {
	id := c.Param("id")
	mu.RLock()
	defer mu.RUnlock()
	for _, inc := range incidents {
		if inc.ID == id {
			c.JSON(200, inc)
			return
		}
	}
	c.JSON(404, gin.H{"error": "incident not found"})
}

func resolveIncident(c *gin.Context) {
	id := c.Param("id")
	mu.Lock()
	for i := range incidents {
		if incidents[i].ID == id {
			incidents[i].Status = "resolved"
			incidents[i].Duration = "resolved"
			mu.Unlock()
			c.JSON(200, incidents[i])
			return
		}
	}
	mu.Unlock()
	c.JSON(404, gin.H{"error": "incident not found"})
}

func getServices(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.JSON(200, gin.H{"services": services, "total": len(services)})
}

func getService(c *gin.Context) {
	name := c.Param("name")
	mu.RLock()
	defer mu.RUnlock()
	for _, s := range services {
		if s.Name == name {
			c.JSON(200, s)
			return
		}
	}
	c.JSON(404, gin.H{"error": "service not found"})
}

func getAlertRules(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.JSON(200, gin.H{"rules": alertRules, "total": len(alertRules)})
}

func toggleAlertRule(c *gin.Context) {
	id := c.Param("id")
	mu.Lock()
	for i := range alertRules {
		if alertRules[i].ID == id {
			alertRules[i].Enabled = !alertRules[i].Enabled
			mu.Unlock()
			c.JSON(200, alertRules[i])
			return
		}
	}
	mu.Unlock()
	c.JSON(404, gin.H{"error": "rule not found"})
}

func getMetricsQuery(c *gin.Context) {
	metric := c.DefaultQuery("metric", "cpu_usage")
	service := c.DefaultQuery("service", "")
	now := time.Now()

	type point struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
		Avg       float64 `json:"avg"`
		P95       float64 `json:"p95"`
		P99       float64 `json:"p99"`
	}

	points := make([]point, 24)
	for i := 0; i < 24; i++ {
		t := now.Add(time.Duration(-23+i) * time.Hour)
		base := 45.0 + rand.Float64()*20
		if metric == "error_rate" {
			base = 0.1 + rand.Float64()*0.5
		} else if metric == "latency" {
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

	c.JSON(200, gin.H{"metric": metric, "service": service, "points": points})
}

func getMetricsNames(c *gin.Context) {
	names := []string{"cpu_usage", "memory_usage", "disk_usage", "error_rate", "latency_p50", "latency_p99", "request_rate", "connection_count", "gc_pause", "thread_count"}
	c.JSON(200, gin.H{"metrics": names})
}

func getTopology(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()

	nodes := make([]TopologyNode, 0, len(topology))
	for _, n := range topology {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	c.JSON(200, gin.H{"nodes": nodes})
}

func getRCA(c *gin.Context) {
	serviceID := c.Param("serviceId")
	mu.RLock()
	defer mu.RUnlock()

	// Simple BFS root cause analysis
	type rcaResult struct {
		Service    string   `json:"service"`
		RootCause  string   `json:"root_cause"`
		Chain      []string `json:"chain"`
		Confidence string   `json:"confidence"`
	}

	node, ok := topology[serviceID]
	if !ok {
		c.JSON(404, gin.H{"error": "service not found"})
		return
	}

	if node.Status == "healthy" {
		c.JSON(200, gin.H{"service": serviceID, "status": "healthy", "message": "No issues detected"})
		return
	}

	// Find root cause by traversing dependencies
	chain := []string{serviceID}
	rootCause := serviceID
	for _, dep := range node.Deps {
		if depNode, ok := topology[dep]; ok && depNode.Status != "healthy" {
			chain = append(chain, dep)
			rootCause = dep
		}
	}

	c.JSON(200, rcaResult{
		Service:    serviceID,
		RootCause:  rootCause,
		Chain:      chain,
		Confidence: "94%",
	})
}

func getInsights(c *gin.Context) {
	insightType := c.DefaultQuery("type", "root-cause")
	mu.RLock()
	defer mu.RUnlock()

	if items, ok := insights[insightType]; ok {
		c.JSON(200, gin.H{"type": insightType, "insights": items})
		return
	}
	c.JSON(200, gin.H{"type": insightType, "insights": []Insight{}})
}

func getIntegrations(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.JSON(200, gin.H{"integrations": integrations, "total": len(integrations)})
}

func getTeam(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.JSON(200, gin.H{"members": members, "total": len(members)})
}

func login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "email and password required"})
		return
	}
	// Demo: accept any login
	c.JSON(200, gin.H{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo-token",
		"user":  gin.H{"id": "U001", "name": "Leo Hang", "email": req.Email, "role": "admin"},
	})
}

// ==================== WebSocket ====================

func handleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
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
	mu.RLock()
	defer mu.RUnlock()
	msg := gin.H{"type": eventType, "data": data, "time": time.Now().UTC()}
	for conn := range clients {
		conn.WriteJSON(msg)
	}
}

// Simulate periodic events
func simulateEvents() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		events := []string{"alert_firing", "incident_update", "service_status"}
		event := events[rand.Intn(len(events))]
		broadcastEvent(event, gin.H{"message": fmt.Sprintf("Simulated %s event", event)})
	}
}

// ==================== Main ====================

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8800"
	}

	initData()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health
	r.GET("/healthz", healthCheck)

	// WebSocket
	r.GET("/api/v1/ws", handleWS)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth
		v1.POST("/auth/login", login)

		// Dashboard
		v1.GET("/dashboard/summary", getDashboardSummary)
		v1.GET("/dashboard/error-rate", getErrorRate)
		v1.GET("/dashboard/latency", getLatency)
		v1.GET("/dashboard/top-errors", getTopErrors)

		// Incidents
		v1.GET("/incidents", getIncidents)
		v1.GET("/incidents/:id", getIncident)
		v1.POST("/incidents/:id/resolve", resolveIncident)

		// Services
		v1.GET("/services", getServices)
		v1.GET("/services/:name", getService)

		// Alert Rules
		v1.GET("/alert-rules", getAlertRules)
		v1.PATCH("/alert-rules/:id/toggle", toggleAlertRule)

		// Metrics
		v1.GET("/metrics/query", getMetricsQuery)
		v1.GET("/metrics/names", getMetricsNames)

		// Topology
		v1.GET("/topology", getTopology)
		v1.GET("/topology/:serviceId/rca", getRCA)

		// Insights
		v1.GET("/insights", getInsights)

		// Integrations
		v1.GET("/integrations", getIntegrations)

		// Team
		v1.GET("/team", getTeam)
	}

	// Start event simulator
	go simulateEvents()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("Opsight API starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
