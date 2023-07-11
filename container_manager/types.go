package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

type Manager struct {
	ctx    context.Context
	client *client.Client
}

// Service

type Service struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Command      []string          `json:"command,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	VolumeMounts []VolumeMount     `json:"volumemounts,omitempty"`
	Networks     []string          `json:"networks,omitempty"`
	Replicas     uint64            `json:"replicas"`
}

type ServiceStatus struct {
	DesiredReplicas  int                 `json:"desiredreplicas"`
	RunningReplicas  int                 `json:"runningreplicas"`
	LastUpdatedAt    string              `json:"lastupdatedat"`
	ServiceUpdateStatus ServiceUpdateStatus `json:"serviceupdatestatus,omitempty"`
}

type ServiceUpdateStatus struct {
	State ServiceUpdateState `json:"state"`
	Message string `json:"message"`
}

type ServiceUpdateState string

const (
	ServiceUpdateStateUpdating ServiceUpdateState = "updating"
	ServiceUpdateStatePaused ServiceUpdateState = "paused"
	ServiceUpdateStateCompleted ServiceUpdateState = "completed"
	ServiceUpdateStateRollbackStarted ServiceUpdateState = "rollback_started"
	ServiceUpdateStateRollbackPaused ServiceUpdateState = "rollback_paused"
	ServiceUpdateStateRollbackCompleted ServiceUpdateState = "rollback_completed"
	ServiceUpdateStateUnknown ServiceUpdateState = "unknown"
)

// Volume Mount
type VolumeMount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"readonly"`
}
