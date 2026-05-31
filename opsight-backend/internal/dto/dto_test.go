package dto

import (
	"testing"

	"opsight-backend/internal/model"
)

func TestFromService(t *testing.T) {
	s := model.Service{
		Name:    "api-gateway",
		Status:  "healthy",
		RPS:     "4200",
		P50:     "12ms",
		P99:     "45ms",
		ErrRate: "0.02%",
		Uptime:  "99.99%",
		Team:    "Platform",
	}

	result := FromService(s)

	if result.Name != s.Name {
		t.Errorf("expected Name %s, got %s", s.Name, result.Name)
	}
	if result.Status != s.Status {
		t.Errorf("expected Status %s, got %s", s.Status, result.Status)
	}
	if result.RPS != s.RPS {
		t.Errorf("expected RPS %s, got %s", s.RPS, result.RPS)
	}
	if result.P50 != s.P50 {
		t.Errorf("expected P50 %s, got %s", s.P50, result.P50)
	}
	if result.P99 != s.P99 {
		t.Errorf("expected P99 %s, got %s", s.P99, result.P99)
	}
	if result.ErrRate != s.ErrRate {
		t.Errorf("expected ErrRate %s, got %s", s.ErrRate, result.ErrRate)
	}
	if result.Uptime != s.Uptime {
		t.Errorf("expected Uptime %s, got %s", s.Uptime, result.Uptime)
	}
	if result.Team != s.Team {
		t.Errorf("expected Team %s, got %s", s.Team, result.Team)
	}
}

func TestFromIncident(t *testing.T) {
	inc := model.Incident{
		ID:       "INC-4221",
		Summary:  "Memory leak in auth-svc",
		Service:  "auth-service",
		Status:   "critical",
		Duration: "14m",
		Time:     "2 min ago",
		Detail:   "auth-svc v2.4.1 disabled session-cache",
	}

	result := FromIncident(inc)

	if result.ID != inc.ID {
		t.Errorf("expected ID %s, got %s", inc.ID, result.ID)
	}
	if result.Summary != inc.Summary {
		t.Errorf("expected Summary %s, got %s", inc.Summary, result.Summary)
	}
	if result.Service != inc.Service {
		t.Errorf("expected Service %s, got %s", inc.Service, result.Service)
	}
	if result.Status != inc.Status {
		t.Errorf("expected Status %s, got %s", inc.Status, result.Status)
	}
	if result.Duration != inc.Duration {
		t.Errorf("expected Duration %s, got %s", inc.Duration, result.Duration)
	}
	if result.Time != inc.Time {
		t.Errorf("expected Time %s, got %s", inc.Time, result.Time)
	}
	if result.Detail != inc.Detail {
		t.Errorf("expected Detail %s, got %s", inc.Detail, result.Detail)
	}
}

func TestFromAlertRule(t *testing.T) {
	rule := model.AlertRule{
		ID:            "AR-001",
		Name:          "High CPU Usage",
		Condition:     "cpu_usage > 85",
		Threshold:     "85%",
		Service:       "All Services",
		Severity:      "warning",
		LastTriggered: "12 min ago",
		Enabled:       true,
		IsAI:          false,
	}

	result := FromAlertRule(rule)

	if result.ID != rule.ID {
		t.Errorf("expected ID %s, got %s", rule.ID, result.ID)
	}
	if result.Name != rule.Name {
		t.Errorf("expected Name %s, got %s", rule.Name, result.Name)
	}
	if !result.Enabled {
		t.Error("expected Enabled to be true")
	}
	if result.IsAI {
		t.Error("expected IsAI to be false")
	}
}

func TestFromAlertRule_Disabled(t *testing.T) {
	rule := model.AlertRule{
		ID:      "AR-010",
		Name:    "DNS Resolution Failure",
		Enabled: false,
		IsAI:    false,
	}

	result := FromAlertRule(rule)

	if result.Enabled {
		t.Error("expected Enabled to be false")
	}
}

func TestFromInsight(t *testing.T) {
	ins := model.Insight{
		Type:       "root-cause",
		Title:      "Memory leak detected",
		Body:       "auth-svc v2.4.1 disabled session-cache",
		Service:    "auth-service",
		Confidence: "97%",
		Time:       "2 min ago",
		Severity:   "critical",
		Related:    "INC-4221",
	}

	result := FromInsight(ins)

	if result.Type != ins.Type {
		t.Errorf("expected Type %s, got %s", ins.Type, result.Type)
	}
	if result.Title != ins.Title {
		t.Errorf("expected Title %s, got %s", ins.Title, result.Title)
	}
	if result.Confidence != ins.Confidence {
		t.Errorf("expected Confidence %s, got %s", ins.Confidence, result.Confidence)
	}
	if result.Service != ins.Service {
		t.Errorf("expected Service %s, got %s", ins.Service, result.Service)
	}
}

func TestFromTopologyNode(t *testing.T) {
	node := model.TopologyNode{
		Name:   "api-gateway",
		Status: "healthy",
		RPS:    "4200",
		P99:    "45ms",
	}

	result := FromTopologyNode(node)

	if result.ID != node.Name {
		t.Errorf("expected ID %s, got %s", node.Name, result.ID)
	}
	if result.Status != node.Status {
		t.Errorf("expected Status %s, got %s", node.Status, result.Status)
	}
}

func TestFromIntegration(t *testing.T) {
	integration := model.Integration{
		ID:         "INT-001",
		Name:       "Slack Alerts",
		Type:       "slack",
		Category:   "Notification",
		Status:     "connected",
		Enabled:    true,
		EventCount: 1247,
	}

	result := FromIntegration(integration)

	if result.ID != integration.ID {
		t.Errorf("expected ID %s, got %s", integration.ID, result.ID)
	}
	if result.Name != integration.Name {
		t.Errorf("expected Name %s, got %s", integration.Name, result.Name)
	}
	if result.Type != integration.Type {
		t.Errorf("expected Type %s, got %s", integration.Type, result.Type)
	}
	if result.Category != integration.Category {
		t.Errorf("expected Category %s, got %s", integration.Category, result.Category)
	}
	if result.Status != integration.Status {
		t.Errorf("expected Status %s, got %s", integration.Status, result.Status)
	}
	if !result.Enabled {
		t.Error("expected Enabled to be true")
	}
	if result.EventCount != integration.EventCount {
		t.Errorf("expected EventCount %d, got %d", integration.EventCount, result.EventCount)
	}
}

func TestFromTeamMember(t *testing.T) {
	member := model.TeamMember{
		ID:    "U001",
		Name:  "Leo Hang",
		Email: "leo@opsight.io",
		Role:  "Admin",
		Team:  "Platform",
	}

	result := FromTeamMember(member)

	if result.ID != member.ID {
		t.Errorf("expected ID %s, got %s", member.ID, result.ID)
	}
	if result.Name != member.Name {
		t.Errorf("expected Name %s, got %s", member.Name, result.Name)
	}
	if result.Email != member.Email {
		t.Errorf("expected Email %s, got %s", member.Email, result.Email)
	}
	if result.Role != member.Role {
		t.Errorf("expected Role %s, got %s", member.Role, result.Role)
	}
	if result.Team != member.Team {
		t.Errorf("expected Team %s, got %s", member.Team, result.Team)
	}
}