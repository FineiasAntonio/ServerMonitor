package model

// SystemInfo holds host metrics
type SystemInfo struct {
	CPUUsage    float64    `json:"cpu_usage_percent"`
	TotalMemory uint64     `json:"total_memory"`
	FreeMemory  uint64     `json:"free_memory"`
	UsedMemory  uint64     `json:"used_memory"`
	Disks       []DiskInfo `json:"disks"`
	HostOS      string     `json:"host_os"`
}

// DiskInfo holds storage metrics
type DiskInfo struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// ContainerSimple holds simplified container info
type ContainerSimple struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
}

// ContainerActionRequest defines the JSON body for container actions
type ContainerActionRequest struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}
