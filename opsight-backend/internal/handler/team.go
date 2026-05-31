package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetTeam(c *gin.Context) {
	var members []model.TeamMember
	database.DB.Find(&members)

	result := make([]dto.TeamMemberDTO, len(members))
	for i, m := range members {
		result[i] = dto.FromTeamMember(m)
	}
	response.Success(c, gin.H{"members": result, "total": len(result)})
}

func CreateTeamMember(c *gin.Context) {
	var m model.TeamMember
	if err := c.ShouldBindJSON(&m); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}
	if m.Name == "" || m.Email == "" {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "name and email are required")
		return
	}
	if m.ID == "" {
		m.ID = fmt.Sprintf("TM-%d", time.Now().UnixNano()%100000)
	}
	if m.Role == "" {
		m.Role = "viewer"
	}
	if err := database.DB.Create(&m).Error; err != nil {
		response.Error(c, http.StatusConflict, response.ErrBadRequest, "member already exists or invalid data")
		return
	}

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "create", "team", m.ID, "Added team member: "+m.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, dto.FromTeamMember(m))
}

func UpdateTeamMember(c *gin.Context) {
	id := c.Param("id")
	var m model.TeamMember
	if err := database.DB.Where("id = ?", id).First(&m).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "member not found")
		return
	}

	var update model.TeamMember
	if err := c.ShouldBindJSON(&update); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrBadRequest, "invalid request body")
		return
	}

	database.DB.Model(&m).Updates(model.TeamMember{
		Name:  update.Name,
		Email: update.Email,
		Role:  update.Role,
		Team:  update.Team,
	})

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "update", "team", id, "Updated team member: "+m.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	database.DB.Where("id = ?", id).First(&m)
	response.Success(c, dto.FromTeamMember(m))
}

func DeleteTeamMember(c *gin.Context) {
	id := c.Param("id")
	var m model.TeamMember
	if err := database.DB.Where("id = ?", id).First(&m).Error; err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "member not found")
		return
	}

	database.DB.Delete(&m)

	userID, email, _ := auth.GetCurrentUser(c)
	audit.Log(userID, email, "delete", "team", id, "Deleted team member: "+m.Name, c.ClientIP(), c.GetHeader("User-Agent"), "success")

	response.Success(c, gin.H{"message": "member deleted"})
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

func GetAuditLogs(c *gin.Context) {
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

func GetAuditStats(c *gin.Context) {
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
