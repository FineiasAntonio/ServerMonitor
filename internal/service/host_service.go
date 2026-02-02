package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"ServerMonitor/internal/model"

	"github.com/creack/pty"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetHostMetrics returns current CPU, Memory, and Disk usage
func GetHostMetrics() model.SystemInfo {
	// Memory
	v, _ := mem.VirtualMemory()

	// CPU (Last second average)
	c, _ := cpu.Percent(time.Second, false)
	cpuVal := 0.0
	if len(c) > 0 {
		cpuVal = c[0]
	}

	// Disks
	var disks []model.DiskInfo
	partitions, _ := disk.Partitions(false)
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, model.DiskInfo{
			Path:        p.Mountpoint,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	return model.SystemInfo{
		CPUUsage:    cpuVal,
		TotalMemory: v.Total,
		FreeMemory:  v.Free,
		UsedMemory:  v.Used,
		Disks:       disks,
		HostOS:      runtime.GOOS,
	}
}

// StartHostConsole spawns a Shell PTY session (Linux only)
func StartHostConsole() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("host console is only supported on Linux")
	}

	c := exec.Command("bash")
	// pty.Start resizes the terminal to a default size (80x24) usually.
	f, err := pty.Start(c)
	if err != nil {
		return nil, err
	}
	return f, nil
}
