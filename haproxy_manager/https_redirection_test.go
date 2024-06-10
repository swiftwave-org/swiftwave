package haproxymanager

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHttpsRedirection(t *testing.T) {
	domainName := "example.com"
	t.Run("Enable HTTPS redirection", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.EnableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error enabling HTTPS redirection: %s", err)
			return
		}
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, "http-request redirect scheme https code 301 if { hdr(host) -i example.com }"), "HTTPS redirection rule should be in config")
	})
	t.Run("Disable HTTPS redirection", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// create rule
		err := haproxyTestManager.EnableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error enabling HTTPS redirection: %s", err)
			return
		}
		// check if rule exists
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, "http-request redirect scheme https code 301 if { hdr(host) -i example.com }"), "HTTPS redirection rule should be in config before start testing")
		// disable rule
		err = haproxyTestManager.DisableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error disabling HTTPS redirection: %s", err)
			return
		}
		// check if rule exists
		config = fetchConfig(transactionId)
		assert.Check(t, !strings.Contains(config, "http-request redirect scheme https code 301 if { hdr(host) -i example.com }"), "HTTPS redirection rule should not be in config after disabling")
	})
	t.Run("Is HTTPS redirection enabled", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// create rule
		err := haproxyTestManager.EnableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error enabling HTTPS redirection: %s", err)
			return
		}
		// check if rule exists
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Contains(config, "http-request redirect scheme https code 301 if { hdr(host) -i example.com }"), "HTTPS redirection rule should be in config before start testing")
		// check if rule is enabled
		index, err := haproxyTestManager.FetchIndexOfHTTPSRedirection(transactionId, domainName)
		assert.Check(t, err == nil, "Error fetching index of HTTPS redirection")
		assert.Check(t, index != -1, "HTTPS redirection should be enabled")
	})
	t.Run("Should not add HTTPS redirection if HTTPS is already enabled", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// create rule
		err := haproxyTestManager.EnableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error enabling HTTPS redirection: %s", err)
			return
		}
		// add again
		err = haproxyTestManager.EnableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error enabling HTTPS redirection: %s", err)
			return
		}
		// check if only one rule exists
		config := fetchConfig(transactionId)
		assert.Check(t, strings.Count(config, "http-request redirect scheme https code 301 if { hdr(host) -i example.com }") == 1, "Only one HTTPS redirection rule for example.com should be in config")
	})
	t.Run("If rule does not exist, deletion should not raise an error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// delete rule
		err := haproxyTestManager.DisableHTTPSRedirection(transactionId, domainName)
		if err != nil {
			t.Errorf("Error disabling HTTPS redirection: %s", err)
			return
		}
		assert.Check(t, err == nil, "Deleting non-existing HTTPS redirection rule should not raise an error")
	})
}
