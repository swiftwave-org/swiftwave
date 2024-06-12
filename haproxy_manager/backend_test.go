package haproxymanager

import (
	"fmt"
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestHTTPBackend(t *testing.T) {
	serviceName := "test-service"
	servicePort := 8080
	serviceReplicas := 3
	backendProtocol := HTTPBackend

	t.Run("is backend exists", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// backend name
		backendName := haproxyTestManager.GenerateBackendName(backendProtocol, serviceName, servicePort)
		// check if backend exists
		isExists, err := haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "backend should not exist [api]")
		// add backend
		_, err = haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, fmt.Sprintf("backend %s", backendName)), "backend name should be in config")
		assert.Check(t, strings.Contains(config, fmt.Sprintf("server-template %s_container- %d %s:%d no-check init-addr none resolvers docker", serviceName, serviceReplicas, serviceName, servicePort)), "server template should be in config")
		assert.Check(t, isExists == true, "backend does not exist [api]")
	})

	t.Run("add backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, strings.Compare(backendName, haproxyTestManager.GenerateBackendName(backendProtocol, serviceName, servicePort)) == 0, "created backend name should match with generated backend name")
		assert.Check(t, strings.Contains(fetchConfig(transactionId), backendName), "backend name should be in config")
	})

	t.Run("add backend with same pre-existing backend", func(t *testing.T) {})

	t.Run("delete backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err := haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		if !isExists {
			t.Fatal("backend has not been created")
		}
		err = haproxyTestManager.DeleteBackend(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "backend should not exist after deletion")
	})

	t.Run("delete backend with non-existing backend name", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.DeleteBackend(transactionId, "non-existing-backend-name")
		assert.Check(t, err == nil, "delete backend with invalid backend name should not return error")
	})

	t.Run("fetch replica count", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		_, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == serviceReplicas, "replicas count should match with expected replicas count")
	})

	t.Run("update replica count", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		_, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 1, "replicas count should be 1 initially")
		err = haproxyTestManager.UpdateBackendReplicas(transactionId, backendProtocol, serviceName, servicePort, 4)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err = haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 4, "replicas count should be 4 after update")
	})
}

func TestTCPBackend(t *testing.T) {
	serviceName := "test-service"
	servicePort := 8080
	serviceReplicas := 3
	backendProtocol := TCPBackend

	t.Run("is backend exists", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// backend name
		backendName := haproxyTestManager.GenerateBackendName(backendProtocol, serviceName, servicePort)
		// check if backend exists
		isExists, err := haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "backend should not exist [api]")
		// add backend
		_, err = haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, fmt.Sprintf("backend %s", backendName)), "backend name should be in config")
		assert.Check(t, strings.Contains(config, fmt.Sprintf("server-template %s_container- %d %s:%d no-check init-addr none resolvers docker", serviceName, serviceReplicas, serviceName, servicePort)), "server template should be in config")
		assert.Check(t, isExists == true, "backend does not exist [api]")
	})

	t.Run("add backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, strings.Compare(backendName, haproxyTestManager.GenerateBackendName(backendProtocol, serviceName, servicePort)) == 0, "created backend name should match with generated backend name")
		assert.Check(t, strings.Contains(fetchConfig(transactionId), backendName), "backend name should be in config")
	})

	t.Run("add backend with same pre-existing backend", func(t *testing.T) {})

	t.Run("delete backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err := haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		if !isExists {
			t.Fatal("backend has not been created")
		}
		err = haproxyTestManager.DeleteBackend(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "backend should not exist after deletion")
	})

	t.Run("delete backend with non-existing backend name", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.DeleteBackend(transactionId, "non-existing-backend-name")
		assert.Check(t, err == nil, "delete backend with invalid backend name should not return error")
	})

	t.Run("fetch replica count", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		_, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == serviceReplicas, "replicas count should match with expected replicas count")
	})

	t.Run("update replica count", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		_, err := haproxyTestManager.AddBackend(transactionId, backendProtocol, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 1, "replicas count should be 1 initially")
		err = haproxyTestManager.UpdateBackendReplicas(transactionId, backendProtocol, serviceName, servicePort, 4)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err = haproxyTestManager.GetReplicaCount(transactionId, backendProtocol, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 4, "replicas count should be 4 after update")
	})
}
