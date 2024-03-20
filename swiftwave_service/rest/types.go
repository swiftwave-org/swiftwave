package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

// Server : hold references to other components of service
type Server struct {
	EchoServer     *echo.Echo
	Config         *config.Config
	ServiceManager *service_manager.ServiceManager
	WorkerManager  *worker.Manager
}

// ResourceStatsData : struct to hold analytics stats data
type ResourceStatsData struct {
	SystemStat   HostResourceStats                `json:"system"`
	ServiceStats map[string]*ServiceResourceStats `json:"services"`
	TimeStamp    uint64                           `json:"timestamp"`
}

// HostResourceStats : struct to hold host resource stats
type HostResourceStats struct {
	CpuUsagePercent uint8       `json:"cpu_used_percent"`
	MemStat         MemoryStat  `json:"memory"`
	DiskStats       []DiskStat  `json:"disks"`
	NetStat         HostNetStat `json:"network"`
}

type DiskStat struct {
	Path       string  `json:"path"`
	MountPoint string  `json:"mount_point"`
	TotalGB    float32 `json:"total_gb"`
	UsedGB     float32 `json:"used_gb"`
}

type MemoryStat struct {
	TotalGB  float32 `json:"total_gb"`
	UsedGB   float32 `json:"used_gb"`
	CachedGB float32 `json:"cached_gb"`
}

type HostNetStat struct {
	SentKB uint64 `json:"sent_kb"`
	RecvKB uint64 `json:"recv_kb"`
}

// ServiceResourceStats : struct to hold service resource stats
type ServiceResourceStats struct {
	CpuUsagePercent uint8          `json:"cpu_used_percent"`
	UsedMemoryMB    uint64         `json:"used_memory_mb"`
	NetStat         ServiceNetStat `json:"network"`
}

type ServiceNetStat struct {
	SentKB uint64 `json:"sent_kb"`
	RecvKB uint64 `json:"recv_kb"`
}
