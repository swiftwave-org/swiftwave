package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
)

// GenerateBackendName : Generate Backend name for HAProxy
func (s Manager) GenerateBackendName(serviceName string, port int) string {
	return "be_" + serviceName + "_" + strconv.Itoa(port)
}

// isBackendExist : Check backend exist in HAProxy configuration
func (s Manager) isBackendExist(backendName string) (bool, error) {
	// Build query parameters
	isBackendExistRequestQueryParams := QueryParameters{}
	// Send request to check if backend exist
	isBackendExistRes, isBackendExistErr := s.getRequest("/services/haproxy/configuration/backends/"+backendName, isBackendExistRequestQueryParams)
	if isBackendExistErr != nil {
		return false, errors.New("failed to check if backend exist")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] isBackendExist:", err)
		}
	}(isBackendExistRes.Body)
	if isBackendExistRes.StatusCode == 404 {
		return false, nil
	} else if isBackendExistRes.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New("failed to check if backend exist")
}

// TODO: add suppport for update, as replicas may change

// AddBackend : Add Backend to HAProxy configuration
// -- Manage server template with backend
func (s Manager) AddBackend(transactionId string, serviceName string, port int, replicas int) (string, error) {
	backendName := s.GenerateBackendName(serviceName, port)
	// Check if backend exist
	isBackendExist, err := s.isBackendExist(backendName)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if isBackendExist {
		return backendName, nil
	}

	// Build query parameters
	addBackendRequestQueryParams := QueryParameters{}
	addBackendRequestQueryParams.add("transaction_id", transactionId)
	// Add backend request body
	addBackendRequestBody := map[string]interface{}{
		"name": backendName,
		"balance": map[string]interface{}{
			"algorithm": "roundrobin",
		},
	}
	addBackendRequestBodyBytes, err := json.Marshal(addBackendRequestBody)
	if err != nil {
		return "", errors.New("failed to marshal add_backend_request_body")
	}
	// Send add backend request
	backendRes, backendErr := s.postRequest("/services/haproxy/configuration/backends", addBackendRequestQueryParams, bytes.NewReader(addBackendRequestBodyBytes))
	if backendErr != nil || !isValidStatusCode(backendRes.StatusCode) {
		return "", errors.New("failed to add backend")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddBackend:", err)
		}
	}(backendRes.Body)

	// Add server template request body
	if replicas <= 0 {
		replicas = 1
	}
	replicasStr := strconv.Itoa(replicas)
	// Server template prefix
	serverTemplatePrefix := serviceName + "_container-"
	// Add server template query parameters
	addServerTemplateRequestQueryParams := QueryParameters{}
	addServerTemplateRequestQueryParams.add("transaction_id", transactionId)
	addServerTemplateRequestQueryParams.add("backend", backendName)
	// Add server template request body
	addServerTemplateRequestBody := map[string]interface{}{
		"prefix":       serverTemplatePrefix,
		"fqdn":         serviceName,
		"port":         port,
		"check":        "disabled",
		"resolvers":    "docker",
		"init-addr":    "none",
		"num_or_range": replicasStr,
	}
	addServerTemplateRequestBodyBytes, err := json.Marshal(addServerTemplateRequestBody)
	if err != nil {
		return "", errors.New("failed to marshal add_server_template_request_body")
	}
	// Send POST request to haproxy to add server
	serverTemplateRes, serverTemplateErr := s.postRequest("/services/haproxy/configuration/server_templates", addServerTemplateRequestQueryParams, bytes.NewReader(addServerTemplateRequestBodyBytes))
	if serverTemplateErr != nil || !isValidStatusCode(serverTemplateRes.StatusCode) {
		return "", errors.New("failed to add server template")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddBackend:", err)
		}
	}(serverTemplateRes.Body)
	return backendName, nil
}

// UpdateBackendReplicas : Update Backend Replicas
// -- Manage server template with backend
func (s Manager) UpdateBackendReplicas(transactionId string, serviceName string, port int, replicas int) error {
	backendName := s.GenerateBackendName(serviceName, port)
	// Check if backend exist
	isBackendExist, err := s.isBackendExist(backendName)
	if err != nil {
		log.Println(err)
		return err
	}

	if !isBackendExist {
		return errors.New("backend does not exist")
	}

	// Add server template request body
	if replicas <= 0 {
		replicas = 1
	}
	replicasStr := strconv.Itoa(replicas)
	// Server template prefix
	serverTemplatePrefix := serviceName + "_container-"
	// Add template query parameters
	addServerTemplateRequestQueryParams := QueryParameters{}
	addServerTemplateRequestQueryParams.add("transaction_id", transactionId)
	addServerTemplateRequestQueryParams.add("backend", backendName)
	// Add server template request body
	addServerTemplateRequestBody := map[string]interface{}{
		"prefix":       serverTemplatePrefix,
		"fqdn":         serviceName,
		"port":         port,
		"check":        "disabled",
		"resolvers":    "docker",
		"init-addr":    "none",
		"num_or_range": replicasStr,
	}
	addServerTemplateRequestBodyBytes, err := json.Marshal(addServerTemplateRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_server_template_request_body")
	}
	// Send POST request to haproxy to add server
	serverTemplateRes, serverTemplateErr := s.putRequest("/services/haproxy/configuration/server_templates/"+serverTemplatePrefix, addServerTemplateRequestQueryParams, bytes.NewReader(addServerTemplateRequestBodyBytes))
	if serverTemplateErr != nil || !isValidStatusCode(serverTemplateRes.StatusCode) {
		return errors.New("failed to add server template")
	}
	return nil
}

// DeleteBackend Delete Backend from HAProxy configuration
func (s Manager) DeleteBackend(transactionId string, backendName string) error {
	// Build query parameters
	addBackendRequestQueryParams := QueryParameters{}
	addBackendRequestQueryParams.add("transaction_id", transactionId)
	// Send request to delete backend from HAProxy
	backendRes, backendErr := s.deleteRequest("/services/haproxy/configuration/backends/"+backendName, addBackendRequestQueryParams)
	if backendErr != nil {
		return errors.New("failed to delete backend")
	}
	if backendRes.StatusCode == 404 {
		return nil
	} else if !isValidStatusCode(backendRes.StatusCode) {
		return errors.New("failed to delete backend")
	}
	return nil
}
