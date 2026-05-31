package dto

// ServiceDTO is the API response shape for a monitored service.
type ServiceDTO struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	RPS     string   `json:"rps"`
	P50     string   `json:"p50"`
	P99     string   `json:"p99"`
	ErrRate string   `json:"err_rate"`
	Uptime  string   `json:"uptime"`
	Team    string   `json:"team"`
	Deps    []string `json:"deps"`
}

// IncidentDTO is the API response shape for an incident.
type IncidentDTO struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Service  string `json:"service"`
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Time     string `json:"time"`
	Detail   string `json:"detail,omitempty"`
}

// AlertRuleDTO is the API response shape for an alert rule.
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

// InsightDTO is the API response shape for an AI insight.
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

// TopologyNodeDTO is the API response shape for a topology node.
type TopologyNodeDTO struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	RPS    string   `json:"rps"`
	P99    string   `json:"p99"`
	Deps   []string `json:"deps"`
}

// IntegrationDTO is the API response shape for an integration.
type IntegrationDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	Enabled    bool   `json:"enabled"`
	EventCount int    `json:"event_count"`
}

// TeamMemberDTO is the API response shape for a team member.
type TeamMemberDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Team  string `json:"team"`
}

// TopErrorDTO is the API response shape for a top error.
type TopErrorDTO struct {
	Error   string `json:"error"`
	Count   int    `json:"count"`
	Trend   string `json:"trend"`
	Service string `json:"service"`
}
