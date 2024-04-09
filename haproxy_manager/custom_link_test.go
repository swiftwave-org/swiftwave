package haproxymanager

import (
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestCustomLink(t *testing.T) {

	t.Run("add tcp link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend " + haproxyTestManager.GenerateBackendName(TCPBackend, "dummy_backend", 8080)

		// create tcp frontend
		backendName, err := haproxyTestManager.AddBackend(transactionId, TCPBackend, "dummy_backend", 8080, 2)
		if err != nil {
			t.Fatal(err)
		}
		// ensure tcp link doesn't exist
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "tcp link should not exist")
		// add tcp link
		err = haproxyTestManager.AddTCPLink(transactionId, backendName, 8080, "", TCPMode, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// check now
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "tcp link should exist")
	})

	t.Run("add duplicate tcp link should silently ignore", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend " + haproxyTestManager.GenerateBackendName(TCPBackend, "dummy_backend", 8080)

		// create tcp frontend
		backendName, err := haproxyTestManager.AddBackend(transactionId, TCPBackend, "dummy_backend", 8080, 2)
		if err != nil {
			t.Fatal(err)
		}
		// add tcp link
		err = haproxyTestManager.AddTCPLink(transactionId, backendName, 8080, "", TCPMode, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// ensure tcp link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "tcp link should exist")
		// add duplicate tcp link
		err = haproxyTestManager.AddTCPLink(transactionId, backendName, 8080, "", TCPMode, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// ensure tcp link exists only once
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Count(config, output), 1, "tcp link should exist")
	})

	t.Run("delete tcp link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend " + haproxyTestManager.GenerateBackendName(TCPBackend, "dummy_backend", 8080)

		// create tcp frontend
		backendName, err := haproxyTestManager.AddBackend(transactionId, TCPBackend, "dummy_backend", 8080, 2)
		if err != nil {
			t.Fatal(err)
		}
		// add tcp link
		err = haproxyTestManager.AddTCPLink(transactionId, backendName, 8080, "", TCPMode, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// ensure tcp link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "tcp link should exist")
		// delete tcp link
		err = haproxyTestManager.DeleteTCPLink(transactionId, backendName, 8080, "", TCPMode)
		if err != nil {
			t.Fatal(err)
		}
		// ensure tcp link doesn't exist
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "tcp link should not exist")
	})

	t.Run("delete non-existing tcp link should silently ignore [no frontend]", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		// delete tcp link
		err := haproxyTestManager.DeleteTCPLink(transactionId, "dummy_backend", 8080, "", TCPMode)
		assert.NilError(t, err, "delete non-existing tcp link should not return error")
	})
}
