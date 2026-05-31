package handler

import (
	"net/http"
	"sort"

	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetTopology(c *gin.Context) {
	var nodes []model.TopologyNode
	database.DB.Find(&nodes)

	result := make([]dto.TopologyNodeDTO, len(nodes))
	for i, n := range nodes {
		result[i] = dto.FromTopologyNode(n)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	response.Success(c, gin.H{"nodes": result})
}

func GetRCA(c *gin.Context) {
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

	var deps []model.TopologyDependency
	database.DB.Where("node_name = ?", serviceID).Find(&deps)
	for _, dep := range deps {
		var depNode model.TopologyNode
		if err := database.DB.Where("name = ?", dep.DepNodeID).First(&depNode).Error; err == nil && depNode.Status != "healthy" {
			chain = append(chain, dep.DepNodeID)
			rootCause = dep.DepNodeID
		}
	}

	response.Success(c, rcaResult{
		Service:    serviceID,
		RootCause:  rootCause,
		Chain:      chain,
		Confidence: "94%",
	})
}
