package service

import (
	"time"

	"ServerMonitor/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetHostMetrics returns current CPU and Memory usage
func GetHostMetrics() model.SystemInfo {
	// Memory
	v, _ := mem.VirtualMemory()

	// CPU (Last second average)
	c, _ := cpu.Percent(time.Second, false)
	cpuVal := 0.0
	if len(c) > 0 {
		cpuVal = c[0]
	}

	return model.SystemInfo{
		CPUUsage:    cpuVal,
		TotalMemory: v.Total,
		FreeMemory:  v.Free,
		UsedMemory:  v.Used,
	}
}
