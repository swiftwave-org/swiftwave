package haproxymanager

import (
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestPredefinedLink(t *testing.T) {

	t.Run("add http link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"
		// ensure http link doesn't exist
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "http link should not exist")
		// add http link
		err := haproxyTestManager.AddHTTPLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure http link exists
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "http link should exist")
	})

	t.Run("add duplicate http link should silently ignore", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"
		// add http link
		err := haproxyTestManager.AddHTTPLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure http link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "http link should exist")
		// add duplicate http link
		err = haproxyTestManager.AddHTTPLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure http link exists
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Count(config, output), 1, "http link should exist")
	})

	t.Run("delete http link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"

		// add http link
		err := haproxyTestManager.AddHTTPLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure http link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "http link should exist")
		// delete http link
		err = haproxyTestManager.DeleteHTTPLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure http link doesn't exist
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "http link should not exist")
	})

	t.Run("delete non-existing http link should not return error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.DeleteHTTPLink(transactionId, "dummy_backend", "example.com")
		assert.Equal(t, err, nil, "delete non-existing http link should not return error")
	})

	t.Run("add https link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"
		// ensure https link doesn't exist
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "https link should not exist")
		// add https link
		err := haproxyTestManager.AddHTTPSLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure https link exists
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "https link should exist")
	})

	t.Run("add duplicate https link should silently ignore", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"
		// add https link
		err := haproxyTestManager.AddHTTPSLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure https link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "https link should exist")
		// add duplicate https link
		err = haproxyTestManager.AddHTTPSLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure https link exists
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Count(config, output), 1, "https link should exist")
	})

	t.Run("delete https link", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		output := "use_backend dummy_backend if { hdr(host) -i example.com }"

		// add https link
		err := haproxyTestManager.AddHTTPSLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure https link exists
		config := fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), true, "https link should exist")
		// delete https link
		err = haproxyTestManager.DeleteHTTPSLink(transactionId, "dummy_backend", "example.com")
		if err != nil {
			t.Fatal(err)
		}
		// ensure https link doesn't exist
		config = fetchConfig(transactionId)
		assert.Equal(t, strings.Contains(config, output), false, "https link should not exist")
	})

	t.Run("delete non-existing https link should not return error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.DeleteHTTPSLink(transactionId, "dummy_backend", "example.com")
		assert.Equal(t, err, nil, "delete non-existing https link should not return error")
	})

}
