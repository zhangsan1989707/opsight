package collector

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	winNetPrevRecv uint64
	winNetPrevSent uint64
	winNetPrevTime time.Time
	winNetMu       sync.Mutex
)

// collect Windows-specific metric collection using gopsutil.
func collect() (*Metrics, error) {
	hostname, _ := os.Hostname()

	m := &Metrics{
		Hostname:  hostname,
		Timestamp: time.Now().UnixMilli(),
	}

	// CPU
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	cores, _ := cpu.Counts(true)
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}
	m.CPU = CPUMetrics{
		UsagePercent: cpuUsage,
		Cores:        cores,
	}

	// Memory
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	m.Memory = MemMetrics{
		TotalMB:     float64(vmem.Total) / (1024 * 1024),
		UsedMB:      float64(vmem.Used) / (1024 * 1024),
		UsagePercent: vmem.UsedPercent,
	}

	// Disk
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions: %w", err)
	}
	var diskMetrics []DiskMetrics
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		diskMetrics = append(diskMetrics, DiskMetrics{
			MountPoint:   p.Mountpoint,
			TotalMB:      float64(usage.Total) / (1024 * 1024),
			UsedMB:       float64(usage.Used) / (1024 * 1024),
			UsagePercent: usage.UsedPercent,
		})
	}
	m.Disk = diskMetrics

	// Network — calculate per-second rate from cumulative counters.
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		now := time.Now()
		winNetMu.Lock()

		var recvPerSec, sentPerSec float64
		if !winNetPrevTime.IsZero() {
			elapsed := now.Sub(winNetPrevTime).Seconds()
			if elapsed > 0 {
				if netIO[0].BytesRecv >= winNetPrevRecv {
					recvPerSec = float64(netIO[0].BytesRecv-winNetPrevRecv) / elapsed
				}
				if netIO[0].BytesSent >= winNetPrevSent {
					sentPerSec = float64(netIO[0].BytesSent-winNetPrevSent) / elapsed
				}
			}
		}

		winNetPrevRecv = netIO[0].BytesRecv
		winNetPrevSent = netIO[0].BytesSent
		winNetPrevTime = now
		winNetMu.Unlock()

		m.Network = NetMetrics{
			BytesRecvPerSec: recvPerSec,
			BytesSentPerSec: sentPerSec,
		}
	}

	// Load
	loadAvg, err := load.Avg()
	if err == nil {
		m.Load = LoadMetrics{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	return m, nil
}
