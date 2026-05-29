package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"opsight-agent/collector"
)

const (
	agentVersion = "1.0.0"
	reportPath   = "/api/v1/agents/report"
	timeout      = 5 * time.Second
	maxRetries   = 3
)

// ReportPayload is the full payload sent to the Opsight backend.
type ReportPayload struct {
	AgentVersion string              `json:"agent_version"`
	Hostname     string              `json:"hostname"`
	IP           string              `json:"ip"`
	OS           string              `json:"os"`
	Timestamp    int64               `json:"timestamp"`
	CPU          collector.CPUMetrics `json:"cpu"`
	Memory       collector.MemMetrics `json:"memory"`
	Disk         []collector.DiskMetrics `json:"disk"`
	Network      collector.NetMetrics `json:"network"`
	Load         collector.LoadMetrics `json:"load"`
}

// GetVersion returns the agent version string.
func GetVersion() string {
	return agentVersion
}

// Report sends collected metrics to the Opsight backend.
// It retries up to maxRetries times on transient failures.
func Report(serverURL, apiKey string, metrics *collector.Metrics) error {
	payload := ReportPayload{
		AgentVersion: agentVersion,
		Hostname:     metrics.Hostname,
		IP:           getOutboundIP(),
		OS:           runtime.GOOS,
		Timestamp:    metrics.Timestamp,
		CPU:          metrics.CPU,
		Memory:       metrics.Memory,
		Disk:         metrics.Disk,
		Network:      metrics.Network,
		Load:         metrics.Load,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := serverURL + reportPath
	client := &http.Client{Timeout: timeout}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[WARN] report attempt %d/%d failed: %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		// Read and discard response body for connection reuse.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("[INFO] report sent successfully (attempt %d)", attempt)
			return nil
		}

		lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
		log.Printf("[WARN] report attempt %d/%d got status %d", attempt, maxRetries, resp.StatusCode)
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return fmt.Errorf("report failed after %d attempts: %w", maxRetries, lastErr)
}

// getOutboundIP returns the preferred outbound IP address of this host.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return addr.IP.String()
		}
	}
	return "127.0.0.1"
}
