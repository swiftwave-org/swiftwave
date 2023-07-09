package dockermanager

import (
	"context"

	"github.com/docker/docker/client"
)

type Manager struct {
	ctx context.Context
	client *client.Client
}