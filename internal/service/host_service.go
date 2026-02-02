package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

// ListServices returns a list of systemd services (Linux only)
func ListServices() ([]model.ServiceInfo, error) {
	if runtime.GOOS == "windows" {
		return []model.ServiceInfo{}, nil
	}

	// systemctl list-units --type=service --all --no-pager --no-legend
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []model.ServiceInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// Format: UNIT LOAD ACTIVE SUB DESCRIPTION...
		// example: ssh.service loaded active running OpenBSD Secure Shell server
		name := fields[0]
		load := fields[1]
		active := fields[2]
		sub := fields[3]
		desc := ""
		if len(fields) > 4 {
			desc = strings.Join(fields[4:], " ")
		}

		services = append(services, model.ServiceInfo{
			Name:        name,
			LoadState:   load,
			ActiveState: active,
			SubState:    sub,
			Description: desc,
		})
	}
	return services, nil
}

// ControlService sends a command to a systemd service (start/stop/restart)
func ControlService(name, action string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("service control not supported on windows")
	}

	validActions := map[string]bool{"start": true, "stop": true, "restart": true}
	if !validActions[action] {
		return fmt.Errorf("invalid action")
	}

	// Basic injection prevention
	if strings.ContainsAny(name, ";&|") {
		return fmt.Errorf("invalid service name")
	}

	cmd := exec.Command("sudo", "systemctl", action, name)
	return cmd.Run()
}
