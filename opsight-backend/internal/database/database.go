package database

import (
	"fmt"
	"os"
	"time"

	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initialises the PostgreSQL connection with retry and runs AutoMigrate.
func InitDB() *gorm.DB {
	host := envOrDefault("DB_HOST", "postgres")
	port := envOrDefault("DB_PORT", "5432")
	user := envOrDefault("DB_USER", "opsight")
	password := envOrDefault("DB_PASSWORD", "opsight_secret")
	dbName := envOrDefault("DB_NAME", "opsight")
	sslmode := envOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode,
	)

	var db *gorm.DB
	var err error

	for retries := 0; retries < 30; retries++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		logger.Warn().Err(err).Int("retry", retries+1).Msg("Waiting for PostgreSQL...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to PostgreSQL after retries")
		os.Exit(1)
	}

	DB = db
	logger.Info().Str("host", host).Str("db", dbName).Msg("Connected to PostgreSQL")

	err = DB.AutoMigrate(
		&model.User{},
		&model.Service{},
		&model.ServiceDependency{},
		&model.Incident{},
		&model.AlertRule{},
		&model.Insight{},
		&model.TopologyNode{},
		&model.TopologyDependency{},
		&model.Integration{},
		&model.TeamMember{},
		&model.TopError{},
		&model.AuditLog{},
		&model.NotificationChannel{},
		&model.NotificationHistory{},
		&model.AgentInstance{},
		&model.MetricSnapshot{},
		&model.AlertEvent{},
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to run AutoMigrate")
		os.Exit(1)
	}

	logger.Info().Msg("Database migration completed")
	return DB
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
