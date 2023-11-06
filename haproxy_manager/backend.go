package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

// Generate Backend name for HAProxy
func (s Manager) GenerateBackendName(service_name string, port int) string {
	return "be_" + service_name + "_" + strconv.Itoa(port)
}

// Check backend exist in HAProxy configuration
func (s Manager) isBackendExist(backend_name string) (bool, error) {
	// Build query parameterss
	is_backend_exist_request_query_params := QueryParameters{}
	// Send request to check if backend exist
	is_backend_exist_res, is_backend_exist_err := s.getRequest("/services/haproxy/configuration/backends/"+backend_name, is_backend_exist_request_query_params)
	if is_backend_exist_err != nil {
		return false, errors.New("failed to check if backend exist")
	}
	defer is_backend_exist_res.Body.Close()
	if is_backend_exist_res.StatusCode == 404 {
		return false, nil
	} else if is_backend_exist_res.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New("failed to check if backend exist")
}

// TODO: add suppport for update, as replicas may change
// Add Backend to HAProxy configuration
// -- Manage server template with backend
func (s Manager) AddBackend(transaction_id string, service_name string, port int, replicas int) (string, error) {
	backend_name := s.GenerateBackendName(service_name, port)
	// Check if backend exist
	is_backend_exist, err := s.isBackendExist(backend_name)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if is_backend_exist {
		return backend_name, nil
	}

	// Build query parameterss
	add_backend_request_query_params := QueryParameters{}
	add_backend_request_query_params.add("transaction_id", transaction_id)
	// Add backend request body
	add_backend_request_body := map[string]interface{}{
		"name": backend_name,
		"balance": map[string]interface{}{
			"algorithm": "roundrobin",
		},
	}
	add_backend_request_body_bytes, err := json.Marshal(add_backend_request_body)
	if err != nil {
		return "", errors.New("failed to marshal add_backend_request_body")
	}
	// Send add backend request
	backend_res, backend_err := s.postRequest("/services/haproxy/configuration/backends", add_backend_request_query_params, bytes.NewReader(add_backend_request_body_bytes))
	if backend_err != nil || !isValidStatusCode(backend_res.StatusCode) {
		return "", errors.New("failed to add backend")
	}
	defer backend_res.Body.Close()

	// Add server template request body
	if replicas <= 0 {
		replicas = 1
	}
	replicas_str := strconv.Itoa(replicas)
	// Server template prefix
	server_template_prefix := service_name + "_container-"
	// Add seManagerrver template query parameters
	add_server_template_request_query_params := QueryParameters{}
	add_server_template_request_query_params.add("transaction_id", transaction_id)
	add_server_template_request_query_params.add("backend", backend_name)
	// Add server template request body
	add_server_template_request_body := map[string]interface{}{
		"prefix":       server_template_prefix,
		"fqdn":         service_name,
		"port":         port,
		"check":        "disabled",
		"resolvers":    "docker",
		"init-addr":    "libc,none",
		"num_or_range": replicas_str,
	}
	add_server_template_request_body_bytes, err := json.Marshal(add_server_template_request_body)
	if err != nil {
		return "", errors.New("failed to marshal add_server_template_request_body")
	}
	// Send POST request to haproxy to add server
	server_template_res, server_template_err := s.postRequest("/services/haproxy/configuration/server_templates", add_server_template_request_query_params, bytes.NewReader(add_server_template_request_body_bytes))
	if server_template_err != nil || !isValidStatusCode(server_template_res.StatusCode) {
		return "", errors.New("failed to add server template")
	}
	defer server_template_res.Body.Close()
	return backend_name, nil
}

// Delete Backend from HAProxy configuration
func (s Manager) DeleteBackend(transaction_id string, backend_name string) error {
	// Build query parameterss
	add_backend_request_query_params := QueryParameters{}
	add_backend_request_query_params.add("transaction_id", transaction_id)
	// Send request to delete backend from HAProxy
	backend_res, backend_err := s.deleteRequest("/services/haproxy/configuration/backends/"+backend_name, add_backend_request_query_params)
	if backend_err != nil || !isValidStatusCode(backend_res.StatusCode) {
		return errors.New("failed to delete backend")
	}
	defer backend_res.Body.Close()
	return nil
}
