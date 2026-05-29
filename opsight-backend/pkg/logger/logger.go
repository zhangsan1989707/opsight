package logger

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339

	level := os.Getenv("LOG_LEVEL")
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil || level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	var output io.Writer
	if format == "json" {
		output = os.Stdout
	} else {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	Logger = zerolog.New(output).With().Timestamp().Logger()
	// Also set the global logger for convenience
	log.Logger = Logger
}

// Info logs at info level
func Info() *zerolog.Event {
	return log.Info()
}

// Error logs at error level
func Error() *zerolog.Event {
	return log.Error()
}

// Warn logs at warn level
func Warn() *zerolog.Event {
	return log.Warn()
}

// Debug logs at debug level
func Debug() *zerolog.Event {
	return log.Debug()
}

// RequestIDKey is the context key for request ID
const RequestIDKey = "request_id"

// Middleware extracts or generates a request ID and sets it in context
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = c.GetHeader("X-Request-Id")
		}
		if rid == "" {
			rid = generateID()
		}

		c.Set(RequestIDKey, rid)
		c.Header("X-Request-ID", rid)

		c.Next()
	}
}

// GinLogger returns a Gin middleware that logs every request
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		var evt *zerolog.Event
		switch {
		case status >= 500:
			evt = log.Error()
		case status >= 400:
			evt = log.Warn()
		default:
			evt = log.Info()
		}

		rid, _ := c.Get(RequestIDKey)
		evt.Str("request_id", rid.(string)).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("client_ip", clientIP).
			Dur("latency", latency).
			Msg("request")
	}
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
