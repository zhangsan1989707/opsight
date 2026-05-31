package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"opsight-backend/internal/model"
	appLogger "opsight-backend/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

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
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})
		if err == nil {
			break
		}
		appLogger.Warn().Err(err).Int("retry", retries+1).Msg("Waiting for PostgreSQL...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		appLogger.Error().Err(err).Msg("Failed to connect to PostgreSQL after retries")
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error().Err(err).Msg("Failed to get underlying sql.DB")
		os.Exit(1)
	}

	maxOpenConns := envOrDefaultInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := envOrDefaultInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := time.Duration(envOrDefaultInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute
	connMaxIdleTime := time.Duration(envOrDefaultInt("DB_CONN_MAX_IDLE_MINUTES", 10)) * time.Minute

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	DB = db
	appLogger.Info().
		Str("host", host).
		Str("db", dbName).
		Int("max_open_conns", maxOpenConns).
		Int("max_idle_conns", maxIdleConns).
		Msg("Connected to PostgreSQL with connection pool")

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
		appLogger.Error().Err(err).Msg("Failed to run AutoMigrate")
		os.Exit(1)
	}

	appLogger.Info().Msg("Database migration completed")
	return DB
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
