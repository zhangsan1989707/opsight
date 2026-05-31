package dto

import (
	"opsight-backend/internal/model"

	"gorm.io/gorm"
)

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}

func serviceDeps(svcName string) []string {
	if db == nil {
		return nil
	}
	var deps []model.ServiceDependency
	db.Where("service_name = ?", svcName).Find(&deps)
	result := make([]string, len(deps))
	for i, d := range deps {
		result[i] = d.DependencyID
	}
	return result
}

func topologyDeps(nodeName string) []string {
	if db == nil {
		return nil
	}
	var deps []model.TopologyDependency
	db.Where("node_name = ?", nodeName).Find(&deps)
	result := make([]string, len(deps))
	for i, d := range deps {
		result[i] = d.DepNodeID
	}
	return result
}

func FromService(s model.Service) ServiceDTO {
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

func FromIncident(i model.Incident) IncidentDTO {
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

func FromAlertRule(r model.AlertRule) AlertRuleDTO {
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

func FromInsight(i model.Insight) InsightDTO {
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

func FromTopologyNode(n model.TopologyNode) TopologyNodeDTO {
	return TopologyNodeDTO{
		ID:     n.Name,
		Status: n.Status,
		RPS:    n.RPS,
		P99:    n.P99,
		Deps:   topologyDeps(n.Name),
	}
}

func FromIntegration(i model.Integration) IntegrationDTO {
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

func FromTeamMember(m model.TeamMember) TeamMemberDTO {
	return TeamMemberDTO{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
		Role:  m.Role,
		Team:  m.Team,
	}
}

func FromTopError(e model.TopError) TopErrorDTO {
	return TopErrorDTO{
		Error:   e.Error,
		Count:   e.Count,
		Trend:   e.Trend,
		Service: e.Service,
	}
}
