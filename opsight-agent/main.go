package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"opsight-agent/collector"
	"opsight-agent/reporter"
)

func main() {
	// Load configuration.
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("[FATAL] load config: %v", err)
	}

	// Validate required settings.
	if cfg.Server.URL == "" {
		log.Fatal("[FATAL] server.url is required (set in agent.yaml or OPSIGHT_SERVER_URL)")
	}
	if cfg.Server.APIKey == "" {
		log.Fatal("[FATAL] server.api_key is required (set in agent.yaml or OPSIGHT_API_KEY)")
	}

	hostname, _ := os.Hostname()
	log.Printf("[INFO] Opsight Agent v%s starting on host %s", reporter.GetVersion(), hostname)
	log.Printf("[INFO] Server: %s", cfg.Server.URL)
	log.Printf("[INFO] Collection interval: %ds", cfg.Collector.IntervalSeconds)
	log.Printf("[INFO] Collectors: cpu=%v mem=%v disk=%v net=%v load=%v",
		cfg.Collector.CPU, cfg.Collector.Memory, cfg.Collector.Disk,
		cfg.Collector.Network, cfg.Collector.Load)

	// Signal handling for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Collection loop.
	interval := time.Duration(cfg.Collector.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first collection immediately.
	collectAndReport(cfg)

	for {
		select {
		case <-ticker.C:
			collectAndReport(cfg)
		case sig := <-sigCh:
			log.Printf("[INFO] received signal %v, shutting down", sig)
			return
		}
	}
}

func collectAndReport(cfg Config) {
	metrics, err := collector.Collect()
	if err != nil {
		log.Printf("[ERROR] collection failed: %v", err)
		return
	}

	if err := reporter.Report(cfg.Server.URL, cfg.Server.APIKey, metrics); err != nil {
		log.Printf("[ERROR] report failed: %v", err)
	}
}
