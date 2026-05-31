package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	httpRequestsTotal = newCounterVec("opsight_http_requests_total", "Total HTTP requests", "method", "path", "status")
	httpRequestDuration = newHistogramVec("opsight_http_request_duration_seconds", "HTTP request latency", "method", "path")
	alertsFiredTotal = newCounter("opsight_alerts_fired_total", "Total alerts fired")
	alertsResolvedTotal = newCounter("opsight_alerts_resolved_total", "Total alerts resolved")
	agentsOnline = newGauge("opsight_agents_online", "Number of online agents")
	agentsTotal = newGauge("opsight_agents_total", "Total registered agents")
	redisCacheHits = newCounter("opsight_redis_cache_hits_total", "Redis cache hits")
	redisCacheMisses = newCounter("opsight_redis_cache_misses_total", "Redis cache misses")

	mu sync.Mutex
)

type counterVec struct {
	name   string
	help   string
	labels []string
	values map[string]float64
}

type histogramVec struct {
	name    string
	help    string
	labels  []string
	buckets map[string][]float64
}

type counter struct {
	name  string
	help  string
	value float64
}

type gauge struct {
	name  string
	help  string
	value float64
}

func newCounterVec(name, help string, labels ...string) *counterVec {
	return &counterVec{name: name, help: help, labels: labels, values: make(map[string]float64)}
}

func newHistogramVec(name, help string, labels ...string) *histogramVec {
	return &histogramVec{name: name, help: help, labels: labels}
}

func newCounter(name, help string) *counter {
	return &counter{name: name, help: help}
}

func newGauge(name, help string) *gauge {
	return &gauge{name: name, help: help}
}

func (c *counter) Inc() {
	mu.Lock()
	c.value++
	mu.Unlock()
}

func (c *counter) Add(v float64) {
	mu.Lock()
	c.value += v
	mu.Unlock()
}

func (g *gauge) Set(v float64) {
	mu.Lock()
	g.value = v
	mu.Unlock()
}

func (cv *counterVec) WithLabelValues(values ...string) {
	key := values[0]
	for i := 1; i < len(values); i++ {
		key += "_" + values[i]
	}
	mu.Lock()
	cv.values[key]++
	mu.Unlock()
}

func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status))
}

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
		RecordHTTPRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
	}
}

func PrometheusMetrics() string {
	mu.Lock()
	defer mu.Unlock()

	var m string
	for name, help := range map[string]string{
		httpRequestsTotal.name:   httpRequestsTotal.help,
		alertsFiredTotal.name:    alertsFiredTotal.help,
		alertsResolvedTotal.name: alertsResolvedTotal.help,
		agentsOnline.name:        agentsOnline.help,
		agentsTotal.name:         agentsTotal.help,
		redisCacheHits.name:      redisCacheHits.help,
		redisCacheMisses.name:    redisCacheMisses.help,
	} {
		var val float64
		switch name {
		case httpRequestsTotal.name:
			for _, v := range httpRequestsTotal.values {
				val += v
			}
		case alertsFiredTotal.name:
			val = alertsFiredTotal.value
		case alertsResolvedTotal.name:
			val = alertsResolvedTotal.value
		case agentsOnline.name:
			val = agentsOnline.value
		case agentsTotal.name:
			val = agentsTotal.value
		case redisCacheHits.name:
			val = redisCacheHits.value
		case redisCacheMisses.name:
			val = redisCacheMisses.value
		}
		m += "# HELP " + name + " " + help + "\n"
		m += "# TYPE " + name + " gauge\n"
		m += name + " " + strconv.FormatFloat(val, 'f', 2, 64) + "\n"
	}

	return m
}
