package dockermanager

import (
	"context"

	"github.com/docker/docker/client"
)

type Manager struct {
	ctx context.Context
	client *client.Client
}

type DockerService struct {
	Name string  `json:"name"`
	Replicas uint64 `json:"replicas"`
	Command []string `json:"command"`
}