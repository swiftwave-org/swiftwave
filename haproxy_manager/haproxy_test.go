package haproxymanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/random"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"testing"
)

// This file contains functions to run integration tests on HAProxy Manager

type haproxyContainer struct {
	testcontainers.Container
	UnixSocketPath string
	VolumeName     string
}

var haproxyTestManager Manager

func TestMain(m *testing.M) {
	//Set up container
	ctx := context.Background()
	haproxyContainer, err := startTestContainer(ctx)
	if err != nil {
		panic(err)
	}
	// Set up haproxy manager
	haproxyTestManager = New(func() (net.Conn, error) {
		return net.Dial("unix", haproxyContainer.UnixSocketPath)
	}, "admin", "admin")
	// Create a transaction id
	//executing all other test suite
	exitCode := m.Run()
	//Destruct database container after completing tests
	if err := haproxyContainer.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}
	// Delete volume
	_, err = exec.Command("docker", "volume", "rm", haproxyContainer.VolumeName).Output()
	if err != nil {
		log.Fatalf("failed to delete volume: %s", err)
	}
	os.Exit(exitCode)
}

func startTestContainer(ctx context.Context) (*haproxyContainer, error) {
	volumeName := "haproxy-data-" + random.String(4)
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/swiftwave-org/haproxy:3.0",
		ExposedPorts: []string{},
		Env: map[string]string{
			"ADMIN_USER":                 "admin",
			"ADMIN_PASSWORD":             "admin",
			"SWIFTWAVE_SERVICE_ENDPOINT": "localhost:80",
			"SWIFTWAVE_SERVICE_ADDRESS":  "localhost",
		},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericVolumeMountSource{
					Name: volumeName,
				},
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
		log.Println("Error starting container")
		return nil, err
	}
	unixSocket := fmt.Sprintf("%s/dataplaneapi.sock", "/var/lib/docker/volumes/"+volumeName+"/_data")
	return &haproxyContainer{
		Container:      container,
		UnixSocketPath: unixSocket,
		VolumeName:     volumeName,
	}, nil
}

// Generate a new transaction id
func newTransaction() string {
	transactionId, err := haproxyTestManager.FetchNewTransactionId()
	if err != nil {
		log.Fatal(err)
	}
	return transactionId
}

// Delete a transaction
func deleteTransaction(transactionId string) {
	err := haproxyTestManager.DeleteTransaction(transactionId)
	if err != nil {
		log.Fatal(err)
	}
}

// Fetch HAProxy configuration
func fetchConfig(transactionId string) string {
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	// Send request
	getConfigRes, getConfigErr := haproxyTestManager.getRequest("/services/haproxy/configuration/raw", params)
	if getConfigErr != nil || !isValidStatusCode(getConfigRes.StatusCode) {
		log.Fatal("failed to fetch config")
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(getConfigRes.Body)
	// Parse response
	var config map[string]interface{}
	err := json.NewDecoder(getConfigRes.Body).Decode(&config)
	if err != nil {
		log.Fatal("failed to parse config")
	}
	return config["data"].(string)
}
