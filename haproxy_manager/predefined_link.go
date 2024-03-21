package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

// AddHTTPLink Add HTTP Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPLink(transactionId string, backendName string, domainName string) error {
	frontendName := "fe_http"
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
	// Build query parameters
	fetchBackendSwitchRequestQueryParams := QueryParameters{}
	fetchBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	fetchBackendSwitchRequestQueryParams.add("frontend", frontendName)
	// Fetch backend switch
	backendSwitchRes, backendSwitchErr := s.getRequest("/services/haproxy/configuration/backend_switching_rules", fetchBackendSwitchRequestQueryParams)
	if backendSwitchErr != nil || !isValidStatusCode(backendSwitchRes.StatusCode) {
		return errors.New("failed to fetch backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPLink: ", err)
		}
	}(backendSwitchRes.Body)
	// Parse backend switch
	backendSwitchData := map[string]interface{}{}
	bodyBytes, err := io.ReadAll(backendSwitchRes.Body)
	if err != nil {
		return errors.New("failed to read backend switch response body")
	}
	err = json.Unmarshal(bodyBytes, &backendSwitchData)
	if err != nil {
		return errors.New("failed to parse backend switch response body")
	}
	// Find backend switch
	backendSwitchDataArray := backendSwitchData["data"].([]interface{})
	backendSwitchDataArrayIndex := -1
	for i, backendSwitchDataArrayItem := range backendSwitchDataArray {
		backendSwitchDataArrayItemMap := backendSwitchDataArrayItem.(map[string]interface{})
		if backendSwitchDataArrayItemMap["name"] == backendName && strings.Contains(backendSwitchDataArrayItemMap["cond_test"].(string), domainName) {
			backendSwitchDataArrayIndex = i
			break
		}
	}
	if backendSwitchDataArrayIndex == -1 {
		return errors.New("failed to find backend switch")
	}
	// Build query parameters
	deleteBackendSwitchRequestQueryParams := QueryParameters{}
	deleteBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	deleteBackendSwitchRequestQueryParams.add("frontend", frontendName)

	// Delete backend switch
	deleteBackendSwitchRes, deleteBackendSwitchErr := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backendSwitchDataArrayIndex), deleteBackendSwitchRequestQueryParams)
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
	fetchBackendSwitchRequestQueryParams := QueryParameters{}
	fetchBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	fetchBackendSwitchRequestQueryParams.add("frontend", frontendName)
	// Fetch backend switch
	backendSwitchRes, backendSwitchErr := s.getRequest("/services/haproxy/configuration/backend_switching_rules", fetchBackendSwitchRequestQueryParams)
	if backendSwitchErr != nil || !isValidStatusCode(backendSwitchRes.StatusCode) {
		return errors.New("failed to fetch backend switch")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPSLink: ", err)
		}
	}(backendSwitchRes.Body)
	// Parse backend switch
	backendSwitchData := map[string]interface{}{}
	bodyBytes, err := io.ReadAll(backendSwitchRes.Body)
	if err != nil {
		return errors.New("failed to read backend switch response body")
	}
	err = json.Unmarshal(bodyBytes, &backendSwitchData)
	if err != nil {
		return errors.New("failed to parse backend switch response body")
	}
	// Find backend switch
	backendSwitchDataArray := backendSwitchData["data"].([]interface{})
	backendSwitchDataArrayIndex := -1
	for i, backendSwitchDataArrayItem := range backendSwitchDataArray {
		backendSwitchDataArrayItemMap := backendSwitchDataArrayItem.(map[string]interface{})
		if backendSwitchDataArrayItemMap["name"] == backendName && strings.Contains(backendSwitchDataArrayItemMap["cond_test"].(string), domainName) {
			backendSwitchDataArrayIndex = i
			break
		}
	}
	if backendSwitchDataArrayIndex == -1 {
		return errors.New("failed to find backend switch")
	}
	// Build query parameters
	deleteBackendSwitchRequestQueryParams := QueryParameters{}
	deleteBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
	deleteBackendSwitchRequestQueryParams.add("frontend", frontendName)

	// Delete backend switch
	deleteBackendSwitchRes, deleteBackendSwitchErr := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backendSwitchDataArrayIndex), deleteBackendSwitchRequestQueryParams)
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
