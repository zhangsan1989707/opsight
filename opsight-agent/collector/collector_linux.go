package collector

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func init() {
	// Ensure we only compile this on Linux.
	// The build constraint in the filename already handles this.
}

// cpuSample holds a single snapshot of /proc/stat CPU line.
type cpuSample struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
	steal   uint64
}

func (c cpuSample) total() uint64 {
	return c.user + c.nice + c.system + c.idle + c.iowait + c.irq + c.softirq + c.steal
}

func (c cpuSample) idleTotal() uint64 {
	return c.idle + c.iowait
}

var (
	prevCPUSample cpuSample
	cpuSampleOnce bool
	cpuSampleMu   sync.Mutex

	prevNetRecv uint64
	prevNetSent uint64
	prevNetTime time.Time
	netMu       sync.Mutex

	coresCache int
	coresOnce  sync.Once
)

// collect Linux-specific metric collection from /proc filesystem.
func collect() (*Metrics, error) {
	hostname, _ := os.Hostname()

	m := &Metrics{
		Hostname:  hostname,
		Timestamp: time.Now().UnixMilli(),
	}

	// CPU
	cpu, err := readCPU()
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	m.CPU = cpu

	// Memory
	mem, err := readMemory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	m.Memory = mem

	// Disk
	disk, err := readDisk()
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	m.Disk = disk

	// Network
	m.Network = readNetwork()

	// Load
	load, err := readLoad()
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	m.Load = load

	return m, nil
}

func getCPUCores() int {
	coresOnce.Do(func() {
		coresCache = runtime.NumCPU()
	})
	return coresCache
}

func readCPU() (CPUMetrics, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return CPUMetrics{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return CPUMetrics{}, fmt.Errorf("empty /proc/stat")
	}

	fields := strings.Fields(scanner.Line())
	if len(fields) < 5 || fields[0] != "cpu" {
		return CPUMetrics{}, fmt.Errorf("unexpected /proc/stat format")
	}

	cur := cpuSample{}
	vals := []*uint64{&cur.user, &cur.nice, &cur.system, &cur.idle, &cur.iowait, &cur.irq, &cur.softirq, &cur.steal}
	for i, f := range fields[1:] {
		if i >= len(vals) {
			break
		}
		*vals[i], _ = strconv.ParseUint(f, 10, 64)
	}

	var usagePercent float64
	cpuSampleMu.Lock()
	if cpuSampleOnce {
		totalDelta := cur.total() - prevCPUSample.total()
		idleDelta := cur.idleTotal() - prevCPUSample.idleTotal()
		if totalDelta > 0 {
			usagePercent = float64(totalDelta-idleDelta) / float64(totalDelta) * 100
		}
	}
	prevCPUSample = cur
	cpuSampleOnce = true
	cpuSampleMu.Unlock()

	return CPUMetrics{
		UsagePercent: usagePercent,
		Cores:        getCPUCores(),
	}, nil
}

func readMemory() (MemMetrics, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemMetrics{}, err
	}
	defer f.Close()

	var totalKB, availKB uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			totalKB = parseKB(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			availKB = parseKB(line)
		}
		if totalKB > 0 && availKB > 0 {
			break
		}
	}

	totalMB := float64(totalKB) / 1024
	usedMB := float64(totalKB-availKB) / 1024
	var usagePercent float64
	if totalKB > 0 {
		usagePercent = float64(totalKB-availKB) / float64(totalKB) * 100
	}

	return MemMetrics{
		TotalMB:     totalMB,
		UsedMB:      usedMB,
		UsagePercent: usagePercent,
	}, nil
}

func parseKB(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(fields[1], 10, 64)
	return v
}

func readDisk() ([]DiskMetrics, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var disks []DiskMetrics
	scanner := bufio.NewScanner(f)

	// Virtual filesystems to skip
	skipFS := map[string]bool{
		"proc": true, "sysfs": true, "devtmpfs": true, "devpts": true,
		"tmpfs": true, "cgroup": true, "cgroup2": true, "pstore": true,
		"bpf": true, "securityfs": true, "debugfs": true, "tracefs": true,
		"fusectl": true, "configfs": true, "hugetlbfs": true, "mqueue": true,
		"rpc_pipefs": true, "overlay": true,
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// Skip virtual filesystems and non-device mounts
		if skipFS[fsType] {
			continue
		}
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		availBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := totalBytes - availBytes

		totalMB := float64(totalBytes) / (1024 * 1024)
		usedMB := float64(usedBytes) / (1024 * 1024)
		var usagePercent float64
		if totalBytes > 0 {
			usagePercent = float64(usedBytes) / float64(totalBytes) * 100
		}

		disks = append(disks, DiskMetrics{
			MountPoint:   mountPoint,
			TotalMB:      totalMB,
			UsedMB:       usedMB,
			UsagePercent: usagePercent,
		})
	}

	return disks, nil
}

func readNetwork() NetMetrics {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return NetMetrics{}
	}
	defer f.Close()

	var totalRecv, totalSent uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header lines
		if strings.Contains(line, "Inter-|") || strings.Contains(line, "face |") {
			continue
		}

		// Skip loopback
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		ifName := strings.TrimSpace(line[:idx])
		if ifName == "lo" {
			continue
		}

		fields := strings.Fields(line[idx+1:])
		if len(fields) < 10 {
			continue
		}

		recv, _ := strconv.ParseUint(fields[0], 10, 64)
		sent, _ := strconv.ParseUint(fields[8], 10, 64)
		totalRecv += recv
		totalSent += sent
	}

	now := time.Now()
	netMu.Lock()
	defer netMu.Unlock()

	var recvPerSec, sentPerSec float64
	if !prevNetTime.IsZero() {
		elapsed := now.Sub(prevNetTime).Seconds()
		if elapsed > 0 && totalRecv >= prevNetRecv && totalSent >= prevNetSent {
			recvPerSec = float64(totalRecv-prevNetRecv) / elapsed
			sentPerSec = float64(totalSent-prevNetSent) / elapsed
		}
	}

	prevNetRecv = totalRecv
	prevNetSent = totalSent
	prevNetTime = now

	return NetMetrics{
		BytesRecvPerSec: recvPerSec,
		BytesSentPerSec: sentPerSec,
	}
}

func readLoad() (LoadMetrics, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return LoadMetrics{}, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return LoadMetrics{}, fmt.Errorf("unexpected /proc/loadavg format")
	}

	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return LoadMetrics{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}, nil
}
