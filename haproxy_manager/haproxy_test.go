package haproxymanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"os"
)

// This file contains functions to run integration tests on HAProxy Manager

type haproxyContainer struct {
	testcontainers.Container
	UnixSocketPath string
}

func startTestContainer(ctx context.Context) (*haproxyContainer, error) {
	// get a temp directory
	tmpDir, err := os.MkdirTemp("", "haproxy-manager-test-*")
	if err != nil {
		fmt.Println("Error creating temp directory")
		return nil, err
	}
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/swiftwave-org/haproxy:3.0",
		ExposedPorts: []string{},
		Env: map[string]string{
			"ADMIN_USER":                 "admin",
			"ADMIN_PASSWORD":             "admin",
			"SWIFTWAVE_SERVICE_ENDPOINT": "localhost:80",
		},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{HostPath: tmpDir},
				Target: "/home",
			},
		},
		WaitingFor: wait.ForExec([]string{"ls", "/home/dataplaneapi.sock"}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Println("Error starting container")
		return nil, err
	}
	unixSocket := fmt.Sprintf("%s/dataplaneapi.sock", tmpDir)
	return &haproxyContainer{
		Container:      container,
		UnixSocketPath: unixSocket,
	}, nil
}

// Fetch HAProxy configuration
func (s Manager) fetchConfig(transactionId string) (string, error) {
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	// Send request
	getConfigRes, getConfigErr := s.getRequest("/services/haproxy/configuration/raw", params)
	if getConfigErr != nil || !isValidStatusCode(getConfigRes.StatusCode) {
		return "", getConfigErr
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(getConfigRes.Body)
	// Parse response
	var config map[string]interface{}
	err := json.NewDecoder(getConfigRes.Body).Decode(&config)
	if err != nil {
		return "", err
	}
	return config["data"].(string), nil
}
