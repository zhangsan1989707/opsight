package handler

import (
	"net/http"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

var authSvc = auth.NewAuthService()

func Login(c *gin.Context) {
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

func Register(c *gin.Context) {
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

func GetCurrentUser(c *gin.Context) {
	userID, _, _ := auth.GetCurrentUser(c)
	user, err := authSvc.GetUserByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "user not found")
		return
	}
	response.Success(c, gin.H{"user": user})
}

func RefreshToken(c *gin.Context) {
	userID, email, role := auth.GetCurrentUser(c)
	token, err := auth.GenerateToken(userID, email, role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternalServer, "failed to refresh token")
		return
	}
	response.Success(c, gin.H{"token": token})
}
