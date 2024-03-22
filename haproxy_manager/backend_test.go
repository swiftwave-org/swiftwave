package haproxymanager

import (
	"fmt"
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestBackend(t *testing.T) {
	serviceName := "test-service"
	servicePort := 8080
	serviceReplicas := 3

	t.Run("is backend exists", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// backend name
		backendName := haproxyTestManager.GenerateBackendName(serviceName, servicePort)
		// check if backend exists
		isExists, err := haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "backend already exists [api]")
		// add backend
		_, err = haproxyTestManager.AddBackend(transactionId, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsBackendExist(transactionId, backendName)
		if err != nil {
			t.Fatal(err)
		}
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, fmt.Sprintf("backend %s", backendName)), "backend not found in config")
		assert.Check(t, strings.Contains(config, fmt.Sprintf("server-template %s_container- %d %s:%d no-check init-addr none resolvers docker", serviceName, serviceReplicas, serviceName, servicePort)), "server template not found in config")
		assert.Check(t, isExists == true, "backend does not exist [api]")
	})

	t.Run("add backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, strings.Compare(backendName, haproxyTestManager.GenerateBackendName(serviceName, servicePort)) == 0, "created backend name is not correct")
		assert.Check(t, strings.Contains(fetchConfig(transactionId), backendName), "backend not found in config")
	})

	t.Run("add backend with existing backend", func(t *testing.T) {})

	t.Run("delete backend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		backendName, err := haproxyTestManager.AddBackend(transactionId, serviceName, servicePort, 1)
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
		assert.Check(t, isExists == false, "backend has not been deleted")
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
		_, err := haproxyTestManager.AddBackend(transactionId, serviceName, servicePort, serviceReplicas)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == serviceReplicas, "replicas count does not match")
	})

	t.Run("update replica count", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		_, err := haproxyTestManager.AddBackend(transactionId, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err := haproxyTestManager.GetReplicaCount(transactionId, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 1, "replicas count does not match")
		err = haproxyTestManager.UpdateBackendReplicas(transactionId, serviceName, servicePort, 1)
		if err != nil {
			t.Fatal(err)
		}
		replicas, err = haproxyTestManager.GetReplicaCount(transactionId, serviceName, servicePort)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, replicas == 1, "replicas count does not match")
	})

}
