package router

import (
	"net/http"
	"time"

	"opsight-backend/internal/audit"
	"opsight-backend/internal/auth"
	"opsight-backend/internal/handler"
	"opsight-backend/pkg/logger"
	"opsight-backend/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.Middleware())
	r.Use(logger.GinLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.InputValidation())
	r.Use(middleware.BodySizeLimit())
	r.Use(middleware.GeneralRateLimit())

	return r
}

func SetupCORS(r *gin.Engine, allowedOrigins []string) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func SetupRoutes(r *gin.Engine, allowedOrigins []string) {
	r.GET("/healthz", handler.HealthCheck)

	handler.SetWSOriginCheck(func(req *http.Request) bool {
		origin := req.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	})

	r.GET("/api/v1/ws", handler.HandleWS)

	v1 := r.Group("/api/v1")
	v1.Use(audit.AuditMiddleware())

	setupAuthRoutes(v1)
	setupDashboardRoutes(v1)
	setupIncidentRoutes(v1)
	setupServiceRoutes(v1)
	setupAlertRoutes(v1)
	setupMetricRoutes(v1)
	setupAgentRoutes(v1)
	setupTopologyRoutes(v1)
	setupInsightRoutes(v1)
	setupIntegrationRoutes(v1)
	setupTeamRoutes(v1)
	setupAuditRoutes(v1)
	setupNotificationRoutes(v1)
}

func setupAuthRoutes(v1 *gin.RouterGroup) {
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", middleware.LoginRateLimit(), handler.Login)

		authProtected := authGroup.Group("")
		authProtected.Use(auth.AuthRequired())
		{
			authProtected.POST("/register", auth.RequireRole("admin"), handler.Register)
			authProtected.GET("/me", handler.GetCurrentUser)
			authProtected.POST("/refresh", handler.RefreshToken)
		}
	}

	v1.POST("/agents/report", handler.AgentAPIKeyAuth(), handler.AgentReport)
}

func setupDashboardRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/dashboard/summary", handler.GetDashboardSummary)
		protected.GET("/dashboard/error-rate", handler.GetErrorRate)
		protected.GET("/dashboard/latency", handler.GetLatency)
		protected.GET("/dashboard/top-errors", handler.GetTopErrors)
	}
}

func setupIncidentRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/incidents", handler.GetIncidents)
		protected.GET("/incidents/:id", handler.GetIncident)
		protected.POST("/incidents/:id/resolve", auth.RequireRole("admin", "editor"), handler.ResolveIncident)
	}
}

func setupServiceRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/services", handler.GetServices)
		protected.GET("/services/:name", handler.GetService)
		protected.POST("/services", auth.RequireRole("admin"), handler.CreateService)
		protected.PUT("/services/:name", auth.RequireRole("admin", "editor"), handler.UpdateService)
		protected.DELETE("/services/:name", auth.RequireRole("admin"), handler.DeleteService)
	}
}

func setupAlertRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/alert-rules", handler.GetAlertRules)
		protected.POST("/alert-rules", auth.RequireRole("admin", "editor"), handler.CreateAlertRule)
		protected.PUT("/alert-rules/:id", auth.RequireRole("admin", "editor"), handler.UpdateAlertRule)
		protected.DELETE("/alert-rules/:id", auth.RequireRole("admin"), handler.DeleteAlertRule)
		protected.PATCH("/alert-rules/:id/toggle", auth.RequireRole("admin", "editor"), handler.ToggleAlertRule)

		protected.GET("/alerts/events", handler.ListAlertEvents)
		protected.GET("/alerts/events/:id", handler.GetAlertEvent)
		protected.POST("/alerts/events/:id/resolve", auth.RequireRole("admin", "editor"), handler.ResolveAlertEvent)
	}
}

func setupMetricRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/metrics/query", handler.GetMetricsQuery)
		protected.GET("/metrics/names", handler.GetMetricsNames)
	}
}

func setupAgentRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/agents", auth.RequireRole("admin"), handler.ListAgents)
		protected.GET("/agents/:hostname", auth.RequireRole("admin"), handler.GetAgent)
		protected.GET("/agents/:hostname/metrics", auth.RequireRole("admin"), handler.GetAgentMetrics)
	}
}

func setupTopologyRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/topology", handler.GetTopology)
		protected.GET("/topology/:serviceId/rca", handler.GetRCA)
	}
}

func setupInsightRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/insights", handler.GetInsights)
	}
}

func setupIntegrationRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/integrations", handler.GetIntegrations)
	}
}

func setupTeamRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/team", auth.RequireRole("admin"), handler.GetTeam)
		protected.POST("/team", auth.RequireRole("admin"), handler.CreateTeamMember)
		protected.PUT("/team/:id", auth.RequireRole("admin"), handler.UpdateTeamMember)
		protected.DELETE("/team/:id", auth.RequireRole("admin"), handler.DeleteTeamMember)
	}
}

func setupAuditRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/audit-logs", auth.RequireRole("admin"), handler.GetAuditLogs)
		protected.GET("/audit-logs/stats", handler.GetAuditStats)
	}
}

func setupNotificationRoutes(v1 *gin.RouterGroup) {
	protected := v1.Group("")
	protected.Use(auth.AuthRequired())
	{
		protected.GET("/notifications/channels", handler.ListNotificationChannels)
		protected.POST("/notifications/channels", auth.RequireRole("admin"), handler.CreateNotificationChannel)
		protected.PUT("/notifications/channels/:id", auth.RequireRole("admin"), handler.UpdateNotificationChannel)
		protected.DELETE("/notifications/channels/:id", auth.RequireRole("admin"), handler.DeleteNotificationChannel)
		protected.GET("/notifications/history", handler.GetNotificationHistory)
		protected.POST("/notifications/test/:channelId", handler.TestNotification)
	}
}