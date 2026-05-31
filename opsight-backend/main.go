package main

import (
	"net/http"
	"os"
	"strings"

	"opsight-backend/internal/cache"
	"opsight-backend/internal/database"
	"opsight-backend/internal/dto"
	"opsight-backend/internal/router"
	"opsight-backend/internal/service"
	"opsight-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	if gin.Mode() != gin.DebugMode {
		if os.Getenv("JWT_SECRET") == "" || os.Getenv("JWT_SECRET") == "opsight-jwt-secret-change-in-production" {
			logger.Error().Msg("JWT_SECRET must be set to a non-default value in production")
			os.Exit(1)
		}
		if os.Getenv("DB_PASSWORD") == "" || os.Getenv("DB_PASSWORD") == "opsight_secret" {
			logger.Error().Msg("DB_PASSWORD must be set to a non-default value in production")
			os.Exit(1)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8800"
	}

	db := database.InitDB()
	database.SeedAll(db)
	dto.SetDB(db)

	cache.InitRedis()

	r := router.New()

	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		allowedOrigins = []string{"http://localhost:3800"}
	}
	router.SetupCORS(r, allowedOrigins)

	router.SetupRoutes(r, allowedOrigins)

	go service.StartAlertEvaluator()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	logger.Info().Str("port", port).Msg("Opsight API starting")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Server error")
		os.Exit(1)
	}
}