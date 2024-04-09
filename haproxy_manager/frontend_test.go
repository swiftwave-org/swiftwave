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
		assert.Check(t, strings.Compare(haproxyTestManager.GenerateFrontendName(TCPMode, 8080), "fe_tcp_8080") == 0, "frontend name should be fe_tcp_8080")
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
		assert.Check(t, strings.Compare(haproxyTestManager.GenerateFrontendName(HTTPMode, 8080), "fe_http_8080") == 0, "frontend name should be fe_http_8080")
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
		backendName, err := haproxyTestManager.AddBackend(transactionId, TCPBackend, "service", 8080, 1)
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
		backendName, err := haproxyTestManager.AddBackend(transactionId, HTTPBackend, "service", 8080, 1)
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

	t.Run("delete http frontend with all switching rules delete should delete frontend", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		backend1Name, err := haproxyTestManager.AddBackend(transactionId, HTTPBackend, "service", 8080, 1)
		if err != nil {
			t.Fatal(err)
		}
		backend2Name, err := haproxyTestManager.AddBackend(transactionId, HTTPBackend, "service", 8081, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddBackendSwitch(transactionId, HTTPMode, 8080, backend1Name, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddBackendSwitch(transactionId, HTTPMode, 8080, backend2Name, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// delete first backend switch
		err = haproxyTestManager.DeleteBackendSwitch(transactionId, HTTPMode, 8080, backend1Name, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// try to delete frontend
		err = haproxyTestManager.DeleteFrontend(transactionId, HTTPMode, 8080)
		assert.Equal(t, err, nil, "delete http frontend with switching rules should not return error")
		// check if frontend exists
		isExists, err := haproxyTestManager.IsFrontendExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == true, "frontend should exist as it has one backend switch [api]")
		// delete second backend switch
		err = haproxyTestManager.DeleteBackendSwitch(transactionId, HTTPMode, 8080, backend2Name, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// try to delete frontend
		err = haproxyTestManager.DeleteFrontend(transactionId, HTTPMode, 8080)
		assert.Equal(t, err, nil, "delete http frontend with switching rules should not return error")
		// check if frontend exists
		isExists, err = haproxyTestManager.IsFrontendExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "frontend should not exist as it has no backend switch [api]")
	})

	t.Run("is switching rule exist", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		isExists, err := haproxyTestManager.IsOtherSwitchingRuleExist(transactionId, HTTPMode, 8080)
		if err != nil {
			t.Fatal(err)
		}
		assert.Check(t, isExists == false, "switching rule should not exist [api]")
		backendName, err := haproxyTestManager.AddBackend(transactionId, HTTPBackend, "service", 8080, 1)
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

	t.Run("if http frontend exists can't create tcp frontend at same port", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		assert.Check(t, err != nil, "tcp frontend creation should fail if http frontend exists")
	})

	t.Run("if tcp frontend exists can't create http frontend at same port", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 8080, []int{})
		if err != nil {
			t.Fatal(err)
		}
		err = haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})
		assert.Check(t, err != nil, "http frontend creation should fail if tcp frontend exists")
	})

	t.Run("add tcp frontend at port 80 should raise error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 80, []int{})
		assert.Check(t, err != nil, "tcp frontend creation should fail for port 80")
	})

	t.Run("add tcp frontend at port 443 should raise error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddFrontend(transactionId, TCPMode, 443, []int{})
		assert.Check(t, err != nil, "tcp frontend creation should fail for port 443")
	})
}
