package database

import (
	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// IsEmpty checks whether a table has no rows.
func IsEmpty(db *gorm.DB, model interface{}) bool {
	var count int64
	db.Model(model).Count(&count)
	return count == 0
}

// SeedAll runs every seed function if the corresponding table is empty.
func SeedAll(db *gorm.DB) {
	if IsEmpty(db, &model.User{}) {
		SeedUsers(db)
	}
	if IsEmpty(db, &model.Service{}) {
		SeedServices(db)
	}
	if IsEmpty(db, &model.Incident{}) {
		SeedIncidents(db)
	}
	if IsEmpty(db, &model.AlertRule{}) {
		SeedAlertRules(db)
	}
	if IsEmpty(db, &model.Insight{}) {
		SeedInsights(db)
	}
	if IsEmpty(db, &model.TopologyNode{}) {
		SeedTopology(db)
	}
	if IsEmpty(db, &model.Integration{}) {
		SeedIntegrations(db)
	}
	if IsEmpty(db, &model.TeamMember{}) {
		SeedMembers(db)
	}
	if IsEmpty(db, &model.TopError{}) {
		SeedTopErrors(db)
	}
	if IsEmpty(db, &model.NotificationChannel{}) {
		SeedNotificationChannels(db)
	}
	if IsEmpty(db, &model.AgentInstance{}) {
		SeedAgentInstances(db)
	}
	if IsEmpty(db, &model.AlertEvent{}) {
		SeedAlertEvents(db)
	}
	logger.Info().Msg("Seed data loaded")
}

func SeedUsers(db *gorm.DB) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash admin password")
		return
	}
	db.Create(&model.User{
		Name:  "Leo Hang",
		Email: "admin@opsight.io",
		Password: string(hash),
		Role:  "admin",
		Team:  "Platform",
	})
}

func SeedServices(db *gorm.DB) {
	services := []model.Service{
		{Name: "api-gateway", Status: "healthy", RPS: "4,200", P50: "12ms", P99: "45ms", ErrRate: "0.02%", Uptime: "99.99%", Team: "Platform"},
		{Name: "auth-service", Status: "down", RPS: "1,100", P50: "8.5s", P99: "12s", ErrRate: "14.2%", Uptime: "99.81%", Team: "Identity"},
		{Name: "payment-gateway", Status: "degraded", RPS: "890", P50: "1.2s", P99: "2.1s", ErrRate: "3.8%", Uptime: "99.92%", Team: "Payments"},
		{Name: "user-service", Status: "healthy", RPS: "3,400", P50: "8ms", P99: "32ms", ErrRate: "0.01%", Uptime: "99.98%", Team: "Identity"},
		{Name: "order-service", Status: "healthy", RPS: "2,100", P50: "18ms", P99: "67ms", ErrRate: "0.05%", Uptime: "99.97%", Team: "Commerce"},
		{Name: "notification-svc", Status: "healthy", RPS: "560", P50: "6ms", P99: "28ms", ErrRate: "0.03%", Uptime: "99.95%", Team: "Platform"},
		{Name: "cache-layer", Status: "degraded", RPS: "12,000", P50: "2ms", P99: "210ms", ErrRate: "0.8%", Uptime: "99.88%", Team: "Infrastructure"},
		{Name: "event-pipeline", Status: "healthy", RPS: "8,700", P50: "4ms", P99: "15ms", ErrRate: "0.01%", Uptime: "99.99%", Team: "Data"},
		{Name: "search-service", Status: "healthy", RPS: "1,800", P50: "22ms", P99: "89ms", ErrRate: "0.08%", Uptime: "99.94%", Team: "Discovery"},
		{Name: "cdn-edge", Status: "healthy", RPS: "45,000", P50: "2ms", P99: "8ms", ErrRate: "0.00%", Uptime: "100.00%", Team: "Infrastructure"},
		{Name: "analytics-svc", Status: "healthy", RPS: "2,300", P50: "35ms", P99: "120ms", ErrRate: "0.12%", Uptime: "99.91%", Team: "Data"},
		{Name: "service-mesh", Status: "healthy", RPS: "-", P50: "0.5ms", P99: "2ms", ErrRate: "0.00%", Uptime: "99.99%", Team: "Infrastructure"},
	}
	for _, s := range services {
		db.Create(&s)
	}

	deps := map[string][]string{
		"api-gateway":      {"user-service", "auth-service", "payment-gateway", "order-service", "cache-layer"},
		"auth-service":     {"cache-layer", "user-service"},
		"payment-gateway":  {"cache-layer", "event-pipeline"},
		"user-service":     {"cache-layer"},
		"order-service":    {"payment-gateway", "event-pipeline", "user-service"},
		"notification-svc": {"event-pipeline"},
		"search-service":   {"cache-layer"},
		"analytics-svc":    {"event-pipeline"},
	}
	for svc, depList := range deps {
		for _, dep := range depList {
			db.Create(&model.ServiceDependency{ServiceName: svc, DependencyID: dep})
		}
	}
}

func SeedIncidents(db *gorm.DB) {
	incidents := []model.Incident{
		{ID: "INC-4221", Summary: "Memory leak in auth-svc causing OOM kills", Service: "auth-service", Status: "critical", Duration: "14m", Time: "2 min ago", Detail: "auth-svc v2.4.1 disabled session-cache eviction. Memory grows 12 MB/min."},
		{ID: "INC-4220", Summary: "Elevated 5xx on /api/v2/payments", Service: "payment-gateway", Status: "critical", Duration: "8m", Time: "12 min ago", Detail: "Redis latency spike correlated with payment-gateway 5xx errors after deploy at 14:05 UTC."},
		{ID: "INC-4219", Summary: "Redis cluster latency spike > 200ms", Service: "cache-layer", Status: "warning", Duration: "23m", Time: "28 min ago", Detail: "Shared Redis instance showing disk IO saturation."},
		{ID: "INC-4218", Summary: "Disk usage > 90% on us-east-1 node 7", Service: "infra/storage", Status: "warning", Duration: "1h 12m", Time: "1 hr ago", Detail: "Log rotation disabled after infra update. Growth rate 2.1 GB/hr."},
		{ID: "INC-4217", Summary: "Kafka consumer lag exceeds threshold", Service: "event-pipeline", Status: "resolved", Duration: "45m", Time: "2 hr ago", Detail: "Consumer group scaled from 3 to 6. Lag resolved."},
		{ID: "INC-4216", Summary: "DNS resolution failure for internal registry", Service: "service-mesh", Status: "resolved", Duration: "12m", Time: "3 hr ago", Detail: "CoreDNS OOM killed after memory limit reduction."},
		{ID: "INC-4215", Summary: "SSL certificate expiring on api-gateway", Service: "api-gateway", Status: "info", Duration: "-", Time: "4 hr ago", Detail: "Certificate expires in 14 days."},
		{ID: "INC-4214", Summary: "Pod restart loop in notification-svc", Service: "notification-service", Status: "resolved", Duration: "38m", Time: "5 hr ago", Detail: "Missing env vars after config migration."},
	}
	for _, inc := range incidents {
		db.Create(&inc)
	}
}

func SeedAlertRules(db *gorm.DB) {
	rules := []model.AlertRule{
		{ID: "AR-001", Name: "High CPU Usage", Condition: "cpu_usage > 85", Threshold: "85%", Service: "All Services", Severity: "warning", LastTriggered: "12 min ago", Enabled: true, IsAI: false},
		{ID: "AR-002", Name: "Memory Leak Detection", Condition: "memory_growth > 10MB/min", Threshold: "10 MB/min", Service: "auth-service", Severity: "critical", LastTriggered: "2 min ago", Enabled: true, IsAI: true},
		{ID: "AR-003", Name: "Error Rate Spike", Condition: "error_rate > 5", Threshold: "5%", Service: "All Services", Severity: "critical", LastTriggered: "12 min ago", Enabled: true, IsAI: false},
		{ID: "AR-004", Name: "Latency Anomaly", Condition: "p99 > 3x baseline", Threshold: "3x", Service: "All Services", Severity: "warning", LastTriggered: "28 min ago", Enabled: true, IsAI: true},
		{ID: "AR-005", Name: "Disk Usage Warning", Condition: "disk_usage > 90", Threshold: "90%", Service: "infra/*", Severity: "warning", LastTriggered: "1 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-006", Name: "SSL Certificate Expiry", Condition: "cert_days < 14", Threshold: "14 days", Service: "All Services", Severity: "info", LastTriggered: "4 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-007", Name: "Pod Restart Loop", Condition: "restart_count > 5 in 10m", Threshold: "5/10min", Service: "All Services", Severity: "critical", LastTriggered: "5 hr ago", Enabled: true, IsAI: true},
		{ID: "AR-008", Name: "Kafka Consumer Lag", Condition: "consumer_lag > 10000", Threshold: "10,000", Service: "event-pipeline", Severity: "warning", LastTriggered: "2 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-009", Name: "Connection Pool Exhaustion", Condition: "pool_usage > 90", Threshold: "90%", Service: "payment-gateway", Severity: "critical", LastTriggered: "12 min ago", Enabled: true, IsAI: true},
		{ID: "AR-010", Name: "DNS Resolution Failure", Condition: "dns_fail_rate > 1", Threshold: "1%", Service: "service-mesh", Severity: "critical", LastTriggered: "3 hr ago", Enabled: false, IsAI: false},
		{ID: "AR-011", Name: "Request Timeout Rate", Condition: "timeout_rate > 2", Threshold: "2%", Service: "api-gateway", Severity: "warning", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-012", Name: "Cache Hit Ratio Drop", Condition: "cache_hit < 80", Threshold: "80%", Service: "cache-layer", Severity: "warning", LastTriggered: "28 min ago", Enabled: true, IsAI: true},
		{ID: "AR-013", Name: "Deployment Anomaly", Condition: "post_deploy_error > 2x", Threshold: "2x", Service: "All Services", Severity: "critical", LastTriggered: "12 min ago", Enabled: true, IsAI: true},
		{ID: "AR-014", Name: "Network Saturation", Condition: "bandwidth > 80", Threshold: "80%", Service: "cdn-edge", Severity: "info", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-015", Name: "Service Discovery Failures", Condition: "sd_fail > 0", Threshold: "0", Service: "service-mesh", Severity: "warning", LastTriggered: "3 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-016", Name: "Queue Depth Alert", Condition: "queue_depth > 5000", Threshold: "5,000", Service: "event-pipeline", Severity: "warning", LastTriggered: "2 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-017", Name: "Response Size Anomaly", Condition: "response_kb > 10x avg", Threshold: "10x", Service: "api-gateway", Severity: "info", LastTriggered: "-", Enabled: false, IsAI: true},
		{ID: "AR-018", Name: "GC Pause Time", Condition: "gc_pause > 200ms", Threshold: "200ms", Service: "auth-service", Severity: "warning", LastTriggered: "2 min ago", Enabled: true, IsAI: false},
		{ID: "AR-019", Name: "Thread Pool Saturation", Condition: "thread_usage > 95", Threshold: "95%", Service: "payment-gateway", Severity: "critical", LastTriggered: "12 min ago", Enabled: true, IsAI: true},
		{ID: "AR-020", Name: "Health Check Failures", Condition: "health_fail > 3 consecutive", Threshold: "3", Service: "All Services", Severity: "critical", LastTriggered: "5 hr ago", Enabled: true, IsAI: false},
		{ID: "AR-021", Name: "API Rate Limit Approaching", Condition: "rate_pct > 80", Threshold: "80%", Service: "api-gateway", Severity: "info", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-022", Name: "Database Connection Leak", Condition: "db_conn_growth > 10/hr", Threshold: "10/hr", Service: "user-service", Severity: "warning", LastTriggered: "-", Enabled: false, IsAI: true},
		{ID: "AR-023", Name: "Certificate Chain Invalid", Condition: "cert_valid == false", Threshold: "false", Service: "api-gateway", Severity: "critical", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-024", Name: "Memory Fragmentation", Condition: "mem_frag > 40", Threshold: "40%", Service: "cache-layer", Severity: "info", LastTriggered: "28 min ago", Enabled: true, IsAI: true},
		{ID: "AR-025", Name: "Cross-Region Latency", Condition: "cross_region_p99 > 500ms", Threshold: "500ms", Service: "cdn-edge", Severity: "warning", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-026", Name: "Secret Rotation Overdue", Condition: "secret_age > 90d", Threshold: "90 days", Service: "All Services", Severity: "info", LastTriggered: "-", Enabled: false, IsAI: false},
		{ID: "AR-027", Name: "Retry Storm Detection", Condition: "retry_rate > 30", Threshold: "30%", Service: "order-service", Severity: "warning", LastTriggered: "-", Enabled: true, IsAI: true},
		{ID: "AR-028", Name: "Log Volume Spike", Condition: "log_eps > 50000", Threshold: "50K/s", Service: "All Services", Severity: "info", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-029", Name: "Dependency Health", Condition: "dep_fail > 0", Threshold: "0", Service: "api-gateway", Severity: "critical", LastTriggered: "12 min ago", Enabled: true, IsAI: false},
		{ID: "AR-030", Name: "Anomalous Traffic Pattern", Condition: "traffic_zscore > 3", Threshold: "3σ", Service: "api-gateway", Severity: "warning", LastTriggered: "-", Enabled: true, IsAI: true},
		{ID: "AR-031", Name: "Container Restart Storm", Condition: "restart_storm == true", Threshold: "true", Service: "All Services", Severity: "critical", LastTriggered: "5 hr ago", Enabled: true, IsAI: true},
		{ID: "AR-032", Name: "Config Drift Detected", Condition: "config_hash != expected", Threshold: "match", Service: "All Services", Severity: "warning", LastTriggered: "-", Enabled: false, IsAI: false},
		{ID: "AR-033", Name: "Cost Anomaly", Condition: "hourly_cost > 2x avg", Threshold: "2x", Service: "All Services", Severity: "info", LastTriggered: "-", Enabled: true, IsAI: false},
		{ID: "AR-034", Name: "Data Freshness", Condition: "data_lag > 5min", Threshold: "5 min", Service: "analytics-svc", Severity: "warning", LastTriggered: "-", Enabled: true, IsAI: false},
	}
	for _, r := range rules {
		db.Create(&r)
	}
}

func SeedInsights(db *gorm.DB) {
	insights := []model.Insight{
		{Type: "root-cause", Title: "Memory leak caused by session-cache eviction disabled in auth-svc v2.4.1", Body: "Deploy at 14:05 UTC changed session.cache.eviction.enabled from true to false. Memory grows at 12 MB/min. Affected 3 pods across us-east-1.", Service: "auth-service", Confidence: "97%", Time: "2 min ago", Severity: "critical", Related: "INC-4221"},
		{Type: "root-cause", Title: "Redis latency spike correlated with payment-gateway 5xx errors", Body: "Both anomalies triggered simultaneously after deploy at 14:05 UTC. Redis primary node showing disk IO saturation on the shared cache-layer instance.", Service: "payment-gateway", Confidence: "94%", Time: "12 min ago", Severity: "critical", Related: "INC-4220"},
		{Type: "root-cause", Title: "Kafka consumer lag from scaled-down group not restored", Body: "Consumer group was manually scaled to 3 for maintenance window at 11:00 UTC. Auto-scaler was paused and not re-enabled. Lag grew from 0 to 12,000 over 90 minutes.", Service: "event-pipeline", Confidence: "99%", Time: "2 hr ago", Severity: "warning", Related: "INC-4217"},
		{Type: "root-cause", Title: "DNS resolution failure from CoreDNS OOM kill", Body: "CoreDNS pod was killed by OOM after memory limit reduced from 512Mi to 256Mi during cluster upgrade.", Service: "service-mesh", Confidence: "91%", Time: "3 hr ago", Severity: "info", Related: "INC-4216"},

		{Type: "predictions", Title: "us-east-1 node 7 disk will reach 95% in ~6 hours", Body: "Current growth rate: 2.1 GB/hr. Log rotation disabled after last infra update.", Service: "infra/storage", Confidence: "88%", Time: "1 hr ago", Severity: "warning", Related: "INC-4218"},
		{Type: "predictions", Title: "auth-service failure may cascade to order-service within 30 min", Body: "order-service depends on auth-service for token validation. Current auth-svc error rate of 14.2% will likely exhaust order-service circuit breaker.", Service: "order-service", Confidence: "76%", Time: "5 min ago", Severity: "info", Related: "INC-4221"},
		{Type: "predictions", Title: "SSL certificate for payment-vendor API expires in 14 days", Body: "Certificate for payments-vendor.internal expires on 2026-06-10. Cert-manager has not initiated renewal.", Service: "payment-gateway", Confidence: "100%", Time: "6 hr ago", Severity: "info", Related: ""},

		{Type: "remediation", Title: "Auto-scaled Kafka consumer group from 3 to 6 instances", Body: "Consumer lag dropped from 12,000 to under 100 within 15 minutes.", Service: "event-pipeline", Confidence: "99%", Time: "2 hr ago", Severity: "resolved", Related: "INC-4217"},
		{Type: "remediation", Title: "Auto-restarted CoreDNS with restored memory limit", Body: "Restored memory limit from 256Mi to 512Mi. DNS resolution returned to normal.", Service: "service-mesh", Confidence: "95%", Time: "3 hr ago", Severity: "resolved", Related: "INC-4216"},
		{Type: "remediation", Title: "Triggered log rotation and freed 12 GB on node 7", Body: "Disk usage dropped from 91.2% to 78.9%.", Service: "infra/storage", Confidence: "90%", Time: "1 hr ago", Severity: "resolved", Related: "INC-4218"},
		{Type: "remediation", Title: "Suggested rollback for auth-svc to v2.4.0", Body: "Rollback will re-enable session-cache eviction and resolve the memory leak.", Service: "auth-service", Confidence: "97%", Time: "2 min ago", Severity: "info", Related: "INC-4221"},

		{Type: "patterns", Title: "Shared Redis instance creates blast radius across services", Body: "auth-service, payment-gateway, and user-service all share the same Redis cluster.", Service: "cache-layer", Confidence: "85%", Time: "30 min ago", Severity: "warning", Related: ""},
		{Type: "patterns", Title: "Deploys between 14:00-15:00 UTC have 3x higher incident rate", Body: "67% of incidents coincide with deploys during this window.", Service: "platform", Confidence: "82%", Time: "1 day ago", Severity: "info", Related: ""},
		{Type: "patterns", Title: "notification-svc restarts correlate with config migrations", Body: "3 of the last 4 config migrations triggered restart loops.", Service: "notification-service", Confidence: "78%", Time: "5 hr ago", Severity: "info", Related: "INC-4214"},
	}
	for _, ins := range insights {
		db.Create(&ins)
	}
}

func SeedTopology(db *gorm.DB) {
	nodes := []model.TopologyNode{
		{Name: "api-gateway", Status: "healthy", RPS: "4,200", P99: "45ms"},
		{Name: "auth-service", Status: "down", RPS: "1,100", P99: "12s"},
		{Name: "payment-gateway", Status: "degraded", RPS: "890", P99: "2.1s"},
		{Name: "cache-layer", Status: "degraded", RPS: "12,000", P99: "210ms"},
		{Name: "user-service", Status: "healthy", RPS: "3,400", P99: "32ms"},
		{Name: "notification-svc", Status: "healthy", RPS: "560", P99: "28ms"},
		{Name: "order-service", Status: "healthy", RPS: "2,100", P99: "67ms"},
		{Name: "search-service", Status: "healthy", RPS: "1,800", P99: "89ms"},
		{Name: "analytics-svc", Status: "healthy", RPS: "2,300", P99: "120ms"},
		{Name: "event-pipeline", Status: "healthy", RPS: "8,700", P99: "15ms"},
		{Name: "cdn-edge", Status: "healthy", RPS: "45,000", P99: "8ms"},
		{Name: "service-mesh", Status: "healthy", RPS: "-", P99: "2ms"},
	}
	for _, n := range nodes {
		db.Create(&n)
	}

	topoDeps := map[string][]string{
		"api-gateway":      {"auth-service", "user-service", "payment-gateway", "order-service", "cache-layer"},
		"auth-service":     {"cache-layer", "user-service"},
		"payment-gateway":  {"cache-layer", "event-pipeline"},
		"user-service":     {"cache-layer"},
		"notification-svc": {"event-pipeline"},
		"order-service":    {"payment-gateway", "event-pipeline", "user-service"},
		"search-service":   {"cache-layer"},
		"analytics-svc":    {"event-pipeline"},
	}
	for node, depList := range topoDeps {
		for _, dep := range depList {
			db.Create(&model.TopologyDependency{NodeName: node, DepNodeID: dep})
		}
	}
}

func SeedIntegrations(db *gorm.DB) {
	integrations := []model.Integration{
		{ID: "INT-001", Name: "Slack Alerts", Type: "slack", Category: "Notification", Status: "connected", Enabled: true, EventCount: 1247},
		{ID: "INT-002", Name: "PagerDuty", Type: "pagerduty", Category: "On-call", Status: "connected", Enabled: true, EventCount: 892},
		{ID: "INT-003", Name: "Prometheus", Type: "prometheus", Category: "Metrics", Status: "connected", Enabled: true, EventCount: 45200},
		{ID: "INT-004", Name: "OpenTelemetry", Type: "otel", Category: "Traces", Status: "connected", Enabled: true, EventCount: 23100},
		{ID: "INT-005", Name: "Zabbix", Type: "zabbix", Category: "Legacy Monitoring", Status: "connected", Enabled: true, EventCount: 634},
		{ID: "INT-006", Name: "Grafana", Type: "grafana", Category: "Visualization", Status: "disconnected", Enabled: false, EventCount: 0},
		{ID: "INT-007", Name: "Webhook (Custom)", Type: "webhook", Category: "Custom", Status: "connected", Enabled: true, EventCount: 201},
		{ID: "INT-008", Name: "Filebeat", Type: "filebeat", Category: "Logs", Status: "connected", Enabled: true, EventCount: 89400},
	}
	for _, i := range integrations {
		db.Create(&i)
	}
}

func SeedMembers(db *gorm.DB) {
	members := []model.TeamMember{
		{ID: "U001", Name: "Leo Hang", Email: "leo@opsight.io", Role: "Admin", Team: "Platform"},
		{ID: "U002", Name: "Zhang Wei", Email: "zhangwei@opsight.io", Role: "Editor", Team: "Identity"},
		{ID: "U003", Name: "Li Na", Email: "lina@opsight.io", Role: "Editor", Team: "Payments"},
		{ID: "U004", Name: "Wang Fang", Email: "wangfang@opsight.io", Role: "Viewer", Team: "Commerce"},
		{ID: "U005", Name: "Chen Jie", Email: "chenjie@opsight.io", Role: "Editor", Team: "Infrastructure"},
		{ID: "U006", Name: "Liu Yang", Email: "liuyang@opsight.io", Role: "Editor", Team: "Data"},
		{ID: "U007", Name: "Zhao Min", Email: "zhaomin@opsight.io", Role: "Viewer", Team: "Discovery"},
		{ID: "U008", Name: "Sun Lei", Email: "sunlei@opsight.io", Role: "Editor", Team: "Platform"},
	}
	for _, m := range members {
		db.Create(&m)
	}
}

func SeedTopErrors(db *gorm.DB) {
	errors := []model.TopError{
		{Error: "OutOfMemoryError: Java heap space", Count: 1247, Trend: "up", Service: "auth-service"},
		{Error: "ConnectionPoolTimeoutException", Count: 892, Trend: "up", Service: "payment-gateway"},
		{Error: "RedisTimeoutException: Command timed out", Count: 634, Trend: "stable", Service: "cache-layer"},
		{Error: "KafkaException: OffsetOutOfRange", Count: 201, Trend: "down", Service: "event-pipeline"},
		{Error: "SSLHandshakeException: Remote host closed", Count: 89, Trend: "down", Service: "api-gateway"},
	}
	for _, e := range errors {
		db.Create(&e)
	}
}

func SeedNotificationChannels(db *gorm.DB) {
	channels := []model.NotificationChannel{
		{
			Name:    "邮件通知",
			Type:    "email",
			Config:  `{"smtp_host":"smtp.example.com","smtp_port":"587"}`,
			Enabled: true,
		},
		{
			Name:    "企业微信",
			Type:    "wechat_work",
			Config:  `{"webhook_url":"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"}`,
			Enabled: true,
		},
	}
	for _, ch := range channels {
		db.Create(&ch)
	}
}

// SeedAgentInstances creates placeholder agent instances (empty by default —
// agents register themselves via the report API).
func SeedAgentInstances(db *gorm.DB) {
	// No seed data — agents self-register via POST /api/v1/agents/report
	logger.Info().Msg("Agent instances seeded (empty)")
}

// SeedAlertEvents creates placeholder alert events (empty by default —
// events are created by the evaluateAlerts engine).
func SeedAlertEvents(db *gorm.DB) {
	// No seed data — alert events are generated by evaluateAlerts()
	logger.Info().Msg("Alert events seeded (empty)")
}
