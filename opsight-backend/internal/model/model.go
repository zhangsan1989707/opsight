package model

import "time"

// User represents an Opsight user.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      string    `gorm:"size:50" json:"role"`
	Team      string    `gorm:"size:100" json:"team"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Service represents a monitored microservice.
type Service struct {
	Name      string    `gorm:"primaryKey;size:255" json:"name"`
	Status    string    `gorm:"size:50" json:"status"`
	RPS       string    `gorm:"size:50" json:"rps"`
	P50       string    `gorm:"size:50" json:"p50"`
	P99       string    `gorm:"size:50" json:"p99"`
	ErrRate   string    `gorm:"size:50" json:"err_rate"`
	Uptime    string    `gorm:"size:50" json:"uptime"`
	Team      string    `gorm:"size:100" json:"team"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceDependency records which service depends on which.
type ServiceDependency struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	ServiceName  string `gorm:"size:255;index;not null" json:"service_name"`
	DependencyID string `gorm:"size:255;not null" json:"dependency_id"`
}

// Incident represents an active or resolved incident.
type Incident struct {
	ID        string    `gorm:"primaryKey;size:50" json:"id"`
	Summary   string    `gorm:"size:500;not null" json:"summary"`
	Service   string    `gorm:"size:255" json:"service"`
	Status    string    `gorm:"size:50" json:"status"`
	Duration  string    `gorm:"size:50" json:"duration"`
	Time      string    `gorm:"size:50" json:"time"`
	Detail    string    `gorm:"size:2000" json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertRule defines a monitoring alert rule.
type AlertRule struct {
	ID            string    `gorm:"primaryKey;size:50" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Condition     string    `gorm:"size:500" json:"condition"`
	Threshold     string    `gorm:"size:100" json:"threshold"`
	Service       string    `gorm:"size:255" json:"service"`
	Severity      string    `gorm:"size:50" json:"severity"`
	LastTriggered string    `gorm:"size:50" json:"last_triggered"`
	Enabled       bool      `json:"enabled"`
	IsAI          bool      `json:"is_ai"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Insight stores AI-generated or manual operational insights.
type Insight struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Type       string    `gorm:"size:50;index" json:"type"`
	Title      string    `gorm:"size:500" json:"title"`
	Body       string    `gorm:"size:2000" json:"body"`
	Service    string    `gorm:"size:255" json:"service"`
	Confidence string    `gorm:"size:10" json:"confidence"`
	Time       string    `gorm:"size:50" json:"time"`
	Severity   string    `gorm:"size:50" json:"severity"`
	Related    string    `gorm:"size:50" json:"related,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// TopologyNode represents a node in the service topology graph.
type TopologyNode struct {
	Name      string    `gorm:"primaryKey;size:255" json:"name"`
	Status    string    `gorm:"size:50" json:"status"`
	RPS       string    `gorm:"size:50" json:"rps"`
	P99       string    `gorm:"size:50" json:"p99"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TopologyDependency records edges in the topology graph.
type TopologyDependency struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	NodeName  string `gorm:"size:255;index;not null" json:"node_name"`
	DepNodeID string `gorm:"size:255;not null" json:"dep_node_id"`
}

// Integration represents a third-party tool integration.
type Integration struct {
	ID         string `gorm:"primaryKey;size:50" json:"id"`
	Name       string `gorm:"size:255" json:"name"`
	Type       string `gorm:"size:50" json:"type"`
	Category   string `gorm:"size:50" json:"category"`
	Status     string `gorm:"size:50" json:"status"`
	Enabled    bool   `json:"enabled"`
	EventCount int    `json:"event_count"`
}

// TeamMember represents a member of the platform operations team.
type TeamMember struct {
	ID    string `gorm:"primaryKey;size:50" json:"id"`
	Name  string `gorm:"size:255" json:"name"`
	Email string `gorm:"size:255" json:"email"`
	Role  string `gorm:"size:50" json:"role"`
	Team  string `gorm:"size:100" json:"team"`
}

// TopError records the most frequent application errors.
type TopError struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Error   string `gorm:"size:500" json:"error"`
	Count   int    `json:"count"`
	Trend   string `gorm:"size:50" json:"trend"`
	Service string `gorm:"size:255" json:"service"`
}

// NotificationChannel represents a notification delivery channel.
type NotificationChannel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Type      string    `gorm:"size:50;not null" json:"type"` // email/wechat_work
	Config    string    `gorm:"type:text" json:"config"`       // JSON config
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationHistory records each notification attempt.
type NotificationHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ChannelID   uint      `gorm:"index" json:"channel_id"`
	ChannelName string    `gorm:"size:255" json:"channel_name"`
	AlertRuleID string    `gorm:"size:50;index" json:"alert_rule_id"`
	Severity    string    `gorm:"size:20" json:"severity"`
	Title       string    `gorm:"size:500" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	Status      string    `gorm:"size:20" json:"status"` // success/failed
	Error       string    `gorm:"size:500" json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLog records user actions for compliance auditing.
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	UserName   string    `gorm:"size:255" json:"user_name"`
	Action     string    `gorm:"size:50;index;not null" json:"action"`
	Resource   string    `gorm:"size:100;index" json:"resource"`
	ResourceID string    `gorm:"size:100" json:"resource_id"`
	Detail     string    `gorm:"size:1000" json:"detail"`
	IP         string    `gorm:"size:50" json:"ip"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Status     string    `gorm:"size:20" json:"status"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// AgentInstance represents a registered monitoring agent on a server.
type AgentInstance struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AgentVersion string    `gorm:"size:50" json:"agent_version"`
	Hostname     string    `gorm:"size:255;uniqueIndex" json:"hostname"`
	IP           string    `gorm:"size:50" json:"ip"`
	OS           string    `gorm:"size:100" json:"os"`
	CPUCores     int       `json:"cpu_cores"`
	MemTotalMB   float64   `json:"mem_total_mb"`
	Status       string    `gorm:"size:20;default:'online'" json:"status"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	FirstSeenAt  time.Time `json:"first_seen_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MetricSnapshot stores raw metrics reported by agents.
type MetricSnapshot struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentID      uint      `gorm:"index" json:"agent_id"`
	Hostname     string    `gorm:"size:255;index" json:"hostname"`
	CPUPercent   float64   `json:"cpu_percent"`
	MemTotalMB   float64   `json:"mem_total_mb"`
	MemUsedMB    float64   `json:"mem_used_mb"`
	MemPercent   float64   `json:"mem_percent"`
	DiskTotalMB  float64   `json:"disk_total_mb"`
	DiskUsedMB   float64   `json:"disk_used_mb"`
	DiskPercent  float64   `json:"disk_percent"`
	NetRecvBytes float64   `json:"net_recv_bytes_per_sec"`
	NetSentBytes float64   `json:"net_sent_bytes_per_sec"`
	Load1        float64   `json:"load1"`
	Load5        float64   `json:"load5"`
	Load15       float64   `json:"load15"`
	ReportedAt   time.Time `gorm:"index" json:"reported_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// AlertEvent records actual alert firings.
type AlertEvent struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	AlertRuleID string     `gorm:"size:50;index" json:"alert_rule_id"`
	RuleName    string     `gorm:"size:255" json:"rule_name"`
	Hostname    string     `gorm:"size:255;index" json:"hostname"`
	Severity    string     `gorm:"size:20" json:"severity"`
	Message     string     `gorm:"size:1000" json:"message"`
	MetricValue float64    `json:"metric_value"`
	Threshold   float64    `json:"threshold"`
	Status      string     `gorm:"size:20" json:"status"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
