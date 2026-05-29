package collector

// Metrics holds all collected system metrics for a single sample point.
type Metrics struct {
	Hostname  string        `json:"hostname"`
	Timestamp int64         `json:"timestamp"`
	CPU       CPUMetrics    `json:"cpu"`
	Memory    MemMetrics    `json:"memory"`
	Disk      []DiskMetrics `json:"disk"`
	Network   NetMetrics    `json:"network"`
	Load      LoadMetrics   `json:"load"`
}

// CPUMetrics represents CPU usage statistics.
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

// MemMetrics represents memory usage statistics.
type MemMetrics struct {
	TotalMB     float64 `json:"total_mb"`
	UsedMB      float64 `json:"used_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

// DiskMetrics represents disk usage for a single mount point.
type DiskMetrics struct {
	MountPoint   string  `json:"mount"`
	TotalMB      float64 `json:"total_mb"`
	UsedMB       float64 `json:"used_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetMetrics represents network I/O rates.
type NetMetrics struct {
	BytesRecvPerSec float64 `json:"bytes_recv_per_sec"`
	BytesSentPerSec float64 `json:"bytes_sent_per_sec"`
}

// LoadMetrics represents system load averages.
type LoadMetrics struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// Collect gathers current system metrics. Platform-specific implementations
// are provided in collector_linux.go and collector_windows.go.
func Collect() (*Metrics, error) {
	return collect()
}
