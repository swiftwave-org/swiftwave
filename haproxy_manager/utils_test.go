package haproxymanager

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
)

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
	//executing all other test suite
	exitCode := m.Run()
	//Destruct database container after completing tests
	if err := haproxyContainer.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}
	os.Exit(exitCode)
}

func TestConfigurationEndpoint(t *testing.T) {
	// Creat transaction
	transactionId, err := haproxyTestManager.FetchNewTransactionId()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("TestFetchConfig", func(t *testing.T) {
		// get configuration
		data, err := haproxyTestManager.fetchConfig(transactionId)
		if err != nil {
			t.Fatal(err)
		}
		// data should be more than 0
		if len(data) == 0 {
			t.Error("failed to fetch configuration")
		} else {
			//t.Log(data)
		}
	})
}
