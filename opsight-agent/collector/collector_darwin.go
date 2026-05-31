//go:build darwin

package collector

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	prevNetRecv uint64
	prevNetSent uint64
	prevNetTime time.Time
)

func collect() (*Metrics, error) {
	hostname, _ := os.Hostname()

	m := &Metrics{
		Hostname:  hostname,
		Timestamp: time.Now().UnixMilli(),
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	cores, _ := cpu.Counts(true)
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}
	m.CPU = CPUMetrics{
		UsagePercent: cpuUsage,
		Cores:        cores,
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	m.Memory = MemMetrics{
		TotalMB:      float64(memInfo.Total) / (1024 * 1024),
		UsedMB:       float64(memInfo.Used) / (1024 * 1024),
		UsagePercent: memInfo.UsedPercent,
	}

	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			m.Disk = append(m.Disk, DiskMetrics{
				MountPoint:   p.Mountpoint,
				TotalMB:      float64(usage.Total) / (1024 * 1024),
				UsedMB:       float64(usage.Used) / (1024 * 1024),
				UsagePercent: usage.UsedPercent,
			})
		}
	}

	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		prevNetRecv = netIO[0].BytesRecv
		prevNetSent = netIO[0].BytesSent
		prevNetTime = time.Now()

		time.Sleep(100 * time.Millisecond)

		netIO2, err2 := net.IOCounters(false)
		if err2 == nil && len(netIO2) > 0 {
			elapsed := time.Since(prevNetTime).Seconds()
			if elapsed > 0 {
				recvPerSec := float64(netIO2[0].BytesRecv-prevNetRecv) / elapsed
				sentPerSec := float64(netIO2[0].BytesSent-prevNetSent) / elapsed
				m.Network = NetMetrics{
					BytesRecvPerSec: recvPerSec,
					BytesSentPerSec: sentPerSec,
				}
			}
		}
	}

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