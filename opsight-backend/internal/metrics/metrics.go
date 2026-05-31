package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opsight_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "opsight_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	alertsFiredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opsight_alerts_fired_total",
			Help: "Total number of alerts fired",
		},
	)

	alertsResolvedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opsight_alerts_resolved_total",
			Help: "Total number of alerts resolved",
		},
	)

	agentsOnline = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opsight_agents_online",
			Help: "Number of online agents",
		},
	)

	agentsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opsight_agents_total",
			Help: "Total number of registered agents",
		},
	)

	redisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opsight_redis_cache_hits_total",
			Help: "Total number of Redis cache hits",
		},
	)

	redisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opsight_redis_cache_misses_total",
			Help: "Total number of Redis cache misses",
		},
	)
)

func RecordAlertFired() {
	alertsFiredTotal.Inc()
}

func RecordAlertResolved() {
	alertsResolvedTotal.Inc()
}

func SetAgentsOnline(n int) {
	agentsOnline.Set(float64(n))
}

func SetAgentsTotal(n int) {
	agentsTotal.Set(float64(n))
}

func RecordRedisHit() {
	redisCacheHits.Inc()
}

func RecordRedisMiss() {
	redisCacheMisses.Inc()
}

func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := http.StatusText(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration.Seconds())
	}
}

func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}