package haproxymanager

import (
	"fmt"
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestFrontend(t *testing.T) {
	t.Run("add tcp frontend", func(t *testing.T) {
		// Add frontend
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// Check if frontend exists
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist [api]")
		assert.Check(t, strings.Compare(GenerateFrontendName(TCPMode, 8080), "fe_tcp_8080") == 0, "frontend name should be fe_tcp_8080")
	})

	t.Run("add http frontend", func(t *testing.T) {
		// Add frontend
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// Check if frontend exists
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist [api]")
		assert.Check(t, strings.Compare(GenerateFrontendName(HTTPMode, 8080), "fe_http_8080") == 0, "frontend name should be fe_http_8080")
	})

	t.Run("add duplicate frontend", func(t *testing.T) {
		// Add frontend
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// Add frontend
		err = haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		assert.Check(t, err == nil, "frontend added successfully")
	})

	t.Run("is frontend exist", func(t *testing.T) {
		// Check if frontend exists
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "frontend should not exist [api]")
		// add frontend
		err = haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// Check if frontend exists
		isExists, err = haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist [api]")
	})

	t.Run("delete frontend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		// Check if frontend exists
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist [api]")
		// Delete frontend
		err = haproxyTestManager.DeleteFrontend(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		// Check if frontend exists
		isExists, err = haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "frontend should not exist [api]")
	})

	t.Run("delete non-existing frontend should not return error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.DeleteFrontend(transactionId, TCPMode, 8080)
		assert.Check(t, err == nil, "delete non-existing frontend should not return error")
	})

	t.Run("delete tcp frontend with switching rules should delete frontend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		backendName, err := haproxyTestManager.AddBackend(transactionId, "service", 8080, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddBackendSwitch(transactionId, TCPMode, 8080, backendName, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, strings.Contains(fetchConfig(transactionId), fmt.Sprintf("use_backend %s", backendName)), "for tcp mode, use_backend should be in config")
		err = haproxyTestManager.DeleteFrontend(transactionId, TCPMode, 8080)
		assert.Equal(t, err, nil, "delete rcp frontend with switching rules should not return error")
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, TCPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "frontend should not exist [api]")
	})

	t.Run("delete http frontend with switching rules should not delete frontend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		backendName, err := haproxyTestManager.AddBackend(transactionId, "service", 8080, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddBackendSwitch(transactionId, HTTPMode, 8080, backendName, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		index, err := haproxyTestManager.FetchBackendSwitchIndex(transactionId, HTTPMode, 8080, backendName, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, index != -1, "backend switch should be created")
		err = haproxyTestManager.DeleteFrontend(transactionId, HTTPMode, 8080)
		assert.Equal(t, err, nil, "delete http frontend with switching rules should not return error")
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist [api]")
	})

	t.Run("is switching rule exist", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		isExists, err := haproxyTestManager.IsOtherSwitchingRuleExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "switching rule should not exist [api]")
		backendName, err := haproxyTestManager.AddBackend(transactionId, "service", 8080, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddBackendSwitch(transactionId, HTTPMode, 8080, backendName, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		isExists, err = haproxyTestManager.IsOtherSwitchingRuleExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "switching rule should exist [api]")
	})

	t.Run("creation of frontend should respect restricted ports", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{8080})
		assert.Check(t, err != nil, "frontend creation should fail for restricted port")
	})
}
