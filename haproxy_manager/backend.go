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
func (s Manager) GenerateBackendName(backendProtocol BackendProtocol, serviceName string, port int) string {
	backendName := "be_" + serviceName + "_" + strconv.Itoa(port)
	if backendProtocol == TCPBackend {
		backendName = backendName + "_tcp"
	}
	return backendName
}

// IsBackendExist : Check backend exist in HAProxy configuration
func (s Manager) IsBackendExist(transactionId string, backendName string) (bool, error) {
	// Build query parameters
	isBackendExistRequestQueryParams := QueryParameters{}
	isBackendExistRequestQueryParams.add("transaction_id", transactionId)
	// Send request to check if backend exist
	isBackendExistRes, isBackendExistErr := s.getRequest("/services/haproxy/configuration/backends/"+backendName, isBackendExistRequestQueryParams)
	if isBackendExistErr != nil {
		return false, errors.New("failed to check if backend exist")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] IsBackendExist:", err)
		}
	}(isBackendExistRes.Body)
	if isBackendExistRes.StatusCode == 404 {
		return false, nil
	} else if isBackendExistRes.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New("failed to check if backend exist")
}

// AddBackend : Add Backend to HAProxy configuration
// -- Manage server template with backend
func (s Manager) AddBackend(transactionId string, backendProtocol BackendProtocol, serviceName string, port int, replicas int) (string, error) {
	backendName := s.GenerateBackendName(backendProtocol, serviceName, port)
	// Check if backend exist
	isBackendExist, err := s.IsBackendExist(transactionId, backendName)
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
	if backendProtocol == TCPBackend {
		addBackendRequestBody["mode"] = "tcp"
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

// GetReplicaCount : Fetch Backend Replicas
// -- Manage server template with backend
func (s Manager) GetReplicaCount(transactionId string, backendProtocol BackendProtocol, serviceName string, port int) (int, error) {
	backendName := s.GenerateBackendName(backendProtocol, serviceName, port)
	// Check if backend exist
	isBackendExist, err := s.IsBackendExist(transactionId, backendName)
	if err != nil {
		return 0, err
	}

	if !isBackendExist {
		return 0, errors.New("backend does not exist")
	}

	// Fetch server template request query parameters
	fetchServerTemplateRequestQueryParams := QueryParameters{}
	fetchServerTemplateRequestQueryParams.add("backend", backendName)
	fetchServerTemplateRequestQueryParams.add("transaction_id", transactionId)
	// Send GET request to fetch server template
	serverTemplateRes, serverTemplateErr := s.getRequest("/services/haproxy/configuration/server_templates", fetchServerTemplateRequestQueryParams)
	if serverTemplateErr != nil {
		return 0, errors.New("failed to fetch server template")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] FetchBackendReplicas:", err)
		}
	}(serverTemplateRes.Body)
	if !isValidStatusCode(serverTemplateRes.StatusCode) {
		return 0, errors.New("failed to fetch server template")
	}
	// Read response body
	serverTemplateResBody, err := io.ReadAll(serverTemplateRes.Body)
	if err != nil {
		return 0, errors.New("failed to read server template response body")
	}
	// Unmarshal response body
	serverTemplateResBodyMap := map[string]interface{}{}
	err = json.Unmarshal(serverTemplateResBody, &serverTemplateResBodyMap)
	if err != nil {
		return 0, errors.New("failed to unmarshal server template response body")
	}
	serverTemplates := serverTemplateResBodyMap["data"].([]interface{})
	if len(serverTemplates) == 0 {
		return 0, nil
	}
	// Get server template
	serverTemplate := serverTemplates[0].(map[string]interface{})
	// Get server template replicas
	return strconv.Atoi(serverTemplate["num_or_range"].(string))
}

// UpdateBackendReplicas : Update Backend Replicas
// -- Manage server template with backend
func (s Manager) UpdateBackendReplicas(transactionId string, backendProtocol BackendProtocol, serviceName string, port int, replicas int) error {
	backendName := s.GenerateBackendName(backendProtocol, serviceName, port)
	// Check if backend exist
	isBackendExist, err := s.IsBackendExist(transactionId, backendName)
	if err != nil {
		return err
	}

	if !isBackendExist {
		return errors.New("backend does not exist")
	}

	// Update server template request body
	if replicas <= 0 {
		replicas = 1
	}
	replicasStr := strconv.Itoa(replicas)
	// Server template prefix
	serverTemplatePrefix := serviceName + "_container-"
	// Update template query parameters
	updateServerTemplateRequestQueryParams := QueryParameters{}
	updateServerTemplateRequestQueryParams.add("transaction_id", transactionId)
	updateServerTemplateRequestQueryParams.add("backend", backendName)
	// Update server template request body
	updateServerTemplateRequestBody := map[string]interface{}{
		"prefix":       serverTemplatePrefix,
		"fqdn":         serviceName,
		"port":         port,
		"check":        "disabled",
		"resolvers":    "docker",
		"init-addr":    "none",
		"num_or_range": replicasStr,
	}
	updateServerTemplateRequestBodyBytes, err := json.Marshal(updateServerTemplateRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_server_template_request_body")
	}
	// Send POST request to haproxy to add server
	serverTemplateRes, serverTemplateErr := s.putRequest("/services/haproxy/configuration/server_templates/"+serverTemplatePrefix, updateServerTemplateRequestQueryParams, bytes.NewReader(updateServerTemplateRequestBodyBytes))
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
