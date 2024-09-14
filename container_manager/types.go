package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

type Manager struct {
	ctx    context.Context
	client *client.Client
}

type DeploymentMode string

const (
	DeploymentModeReplicated DeploymentMode = "replicated"
	DeploymentModeGlobal     DeploymentMode = "global"
)

type Service struct {
	Name                 string            `json:"name"`
	Image                string            `json:"image"`
	Hostname             string            `json:"hostname,omitempty"`
	Command              []string          `json:"command,omitempty"`
	Env                  map[string]string `json:"env,omitempty"`
	Capabilities         []string          `json:"capabilities,omitempty"`
	Sysctls              map[string]string `json:"sysctl,omitempty"`
	ConfigMounts         []ConfigMount     `json:"configmounts,omitempty"`
	VolumeMounts         []VolumeMount     `json:"volumemounts,omitempty"`
	VolumeBinds          []VolumeBind      `json:"volumebinds,omitempty"`
	Networks             []string          `json:"networks,omitempty"`
	DeploymentMode       DeploymentMode    `json:"deploymentmode"`
	Replicas             uint64            `json:"replicas"`
	PlacementConstraints []string          `json:"placementconstraints,omitempty"`
	ReservedResource     Resource          `json:"reserved_resource,omitempty"`
	ResourceLimit        Resource          `json:"resource_limit,omitempty"`
	CustomHealthCheck    CustomHealthCheck `json:"custom_health_check,omitempty"`
}

type CustomHealthCheck struct {
	Enabled              bool   `json:"enabled"`
	TestCommand          string `json:"test_command"`
	IntervalSeconds      uint64 `json:"interval_seconds"`       // Time between running the check in seconds
	TimeoutSeconds       uint64 `json:"timeout_seconds"`        // Maximum time to allow one check to run in seconds
	StartPeriodSeconds   uint64 `json:"start_period_seconds"`   // Start period for the container to initialize before counting retries towards unstable
	StartIntervalSeconds uint64 `json:"start_interval_seconds"` // Time between running the check during the start period
	Retries              uint64 `json:"retries"`                // Consecutive failures needed to report unhealthy
}

type ServiceRealtimeInfo struct {
	Name              string `json:"name"`
	DesiredReplicas   int    `json:"desiredreplicas"`
	RunningReplicas   int    `json:"runningreplicas"`
	ReplicatedService bool   `json:"replicatedservice"`
}

type ServiceTaskPlacementInfo struct {
	NodeID          string `json:"nodeid"`
	NodeName        string `json:"nodename"`
	IsManagerNode   bool   `json:"ismanagernode"`
	RunningReplicas int    `json:"runningreplicas"`
}

type VolumeMount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"readonly"`
}

type VolumeBind struct {
	Source string `json:"source"`
	Target string `json:"target"`
	// No need to specify readonly for volume bind
	// VolumeBind is for special internal use, and need readwrite access all the time
}

type ConfigMount struct {
	ConfigID     string `json:"config_id"`
	Uid          uint   `json:"uid"`
	Gid          uint   `json:"gid"`
	FileMode     uint   `json:"file_mode"`
	MountingPath string `json:"mounting_path"`
}

type Resource struct {
	MemoryMB int `json:"memory_mb,omitempty"`
}

type DockerProxyConfig struct {
	Permission DockerProxyPermission `json:"permissions" gorm:"embedded;embeddedPrefix:permission_"`
}

type DockerProxyPermissionType string

const (
	// DockerProxyNoPermission no request will be allowed
	DockerProxyNoPermission DockerProxyPermissionType = "none"
	// DockerProxyReadPermission only [GET, HEAD] requests will be allowed
	DockerProxyReadPermission DockerProxyPermissionType = "read"
	// DockerProxyReadWritePermission all requests will be allowed [GET, HEAD, POST, PUT, DELETE, OPTIONS]
	DockerProxyReadWritePermission DockerProxyPermissionType = "read_write"
)

type DockerProxyPermission struct {
	Ping         DockerProxyPermissionType `json:"ping" gorm:"default:read"`
	Version      DockerProxyPermissionType `json:"version" gorm:"default:none"`
	Info         DockerProxyPermissionType `json:"info" gorm:"default:none"`
	Events       DockerProxyPermissionType `json:"events" gorm:"default:none"`
	Auth         DockerProxyPermissionType `json:"auth" gorm:"default:none"`
	Secrets      DockerProxyPermissionType `json:"secrets" gorm:"default:none"`
	Build        DockerProxyPermissionType `json:"build" gorm:"default:none"`
	Commit       DockerProxyPermissionType `json:"commit" gorm:"default:none"`
	Configs      DockerProxyPermissionType `json:"configs" gorm:"default:none"`
	Containers   DockerProxyPermissionType `json:"containers" gorm:"default:none"`
	Distribution DockerProxyPermissionType `json:"distribution" gorm:"default:none"`
	Exec         DockerProxyPermissionType `json:"exec" gorm:"default:none"`
	Grpc         DockerProxyPermissionType `json:"grpc" gorm:"default:none"`
	Images       DockerProxyPermissionType `json:"images" gorm:"default:none"`
	Networks     DockerProxyPermissionType `json:"networks" gorm:"default:none"`
	Nodes        DockerProxyPermissionType `json:"nodes" gorm:"default:none"`
	Plugins      DockerProxyPermissionType `json:"plugins" gorm:"default:none"`
	Services     DockerProxyPermissionType `json:"services" gorm:"default:none"`
	Session      DockerProxyPermissionType `json:"session" gorm:"default:none"`
	Swarm        DockerProxyPermissionType `json:"swarm" gorm:"default:none"`
	System       DockerProxyPermissionType `json:"system" gorm:"default:none"`
	Tasks        DockerProxyPermissionType `json:"tasks" gorm:"default:none"`
	Volumes      DockerProxyPermissionType `json:"volumes" gorm:"default:none"`
}
