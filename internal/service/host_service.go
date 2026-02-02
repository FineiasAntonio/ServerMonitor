package service

import (
	"time"

	"ServerMonitor/internal/model"

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
	}
}
