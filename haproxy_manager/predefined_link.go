package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
)

// AddHTTPLink Add HTTP Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPLink(transactionId string, backendName string, domainName string) error {
	frontendName := "fe_http"
	// Check if backend switch already exists
	backendSwitchIndex, err := s.FetchBackendSwitchIndexByName(transactionId, frontendName, 80, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex != -1 {
		return nil
	}
	// Build query parameters
	addBackendSwitchRequestQueryParams := QueryParameters{}
	addBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	addBackendSwitchRequestQueryParams.add("frontend", frontendName)
	// Add backend switch request body
	addBackendSwitchRequestBody := map[string]interface{}{
		"cond":      "if",
		"cond_test": `{ hdr(host) -i ` + domainName + ` }`,
		"index":     1,
		"name":      backendName,
	}
	addBackendSwitchRequestBodyBytes, err := json.Marshal(addBackendSwitchRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send add backend switch request
	backendSwitchRes, backendSwitchErr := s.postRequest("/services/haproxy/configuration/backend_switching_rules", addBackendSwitchRequestQueryParams, bytes.NewReader(addBackendSwitchRequestBodyBytes))
	if backendSwitchErr != nil || !isValidStatusCode(backendSwitchRes.StatusCode) {
		return errors.New("failed to add backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddHTTPLink: ", err)
		}
	}(backendSwitchRes.Body)
	return nil
}

// DeleteHTTPLink Delete HTTP Link from HAProxy configuration
func (s Manager) DeleteHTTPLink(transactionId string, backendName string, domainName string) error {
	frontendName := "fe_http"
	backendSwitchIndex, err := s.FetchBackendSwitchIndexByName(transactionId, frontendName, 80, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex == -1 {
		return nil
	}
	// Build query parameters
	deleteBackendSwitchRequestQueryParams := QueryParameters{}
	deleteBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	deleteBackendSwitchRequestQueryParams.add("frontend", frontendName)

	// Delete backend switch
	deleteBackendSwitchRes, deleteBackendSwitchErr := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backendSwitchIndex), deleteBackendSwitchRequestQueryParams)
	if deleteBackendSwitchErr != nil || !isValidStatusCode(deleteBackendSwitchRes.StatusCode) {
		return errors.New("failed to delete backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPLink: ", err)
		}
	}(deleteBackendSwitchRes.Body)
	return nil
}

// AddHTTPSLink Add HTTPS Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPSLink(transactionId string, backendName string, domainName string) error {
	frontendName := "fe_https"
	// Check if backend switch already exists
	backendSwitchIndex, err := s.FetchBackendSwitchIndexByName(transactionId, frontendName, 443, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex != -1 {
		return nil
	}
	// Build query parameters
	addBackendSwitchRequestQueryParams := QueryParameters{}
	addBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	addBackendSwitchRequestQueryParams.add("frontend", frontendName)
	// Add backend switch request body
	addBackendSwitchRequestBody := map[string]interface{}{
		"cond":      "if",
		"cond_test": `{ hdr(host) -i ` + domainName + ` }`,
		"index":     1,
		"name":      backendName,
	}
	addBackendSwitchRequestBodyBytes, err := json.Marshal(addBackendSwitchRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send add backend switch request
	backendSwitchRes, backendSwitchErr := s.postRequest("/services/haproxy/configuration/backend_switching_rules", addBackendSwitchRequestQueryParams, bytes.NewReader(addBackendSwitchRequestBodyBytes))
	if backendSwitchErr != nil || !isValidStatusCode(backendSwitchRes.StatusCode) {
		return errors.New("failed to add backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddHTTPSLink: ", err)
		}
	}(backendSwitchRes.Body)
	return nil
}

// DeleteHTTPSLink Delete HTTPS Link from HAProxy configuration
func (s Manager) DeleteHTTPSLink(transactionId string, backendName string, domainName string) error {
	frontendName := "fe_https"
	// Build query parameters
	backendSwitchIndex, err := s.FetchBackendSwitchIndexByName(transactionId, frontendName, 443, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex == -1 {
		return nil
	}

	// Build query parameters
	deleteBackendSwitchRequestQueryParams := QueryParameters{}
	deleteBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	deleteBackendSwitchRequestQueryParams.add("frontend", frontendName)

	// Delete backend switch
	deleteBackendSwitchRes, deleteBackendSwitchErr := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backendSwitchIndex), deleteBackendSwitchRequestQueryParams)
	if deleteBackendSwitchErr != nil || !isValidStatusCode(deleteBackendSwitchRes.StatusCode) {
		return errors.New("failed to delete backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPSLink: ", err)
		}
	}(deleteBackendSwitchRes.Body)
	return nil
}
