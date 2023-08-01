package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

// Add Backend to HAProxy configuration
// -- Manage server template with backend
func (s Manager) AddBackend(transaction_id string, service_name string, port int, replicas int) error {
	backend_name := "be_" + service_name + "_" + strconv.Itoa(port)
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
		return errors.New("failed to marshal add_backend_request_body")
	}
	// Send add backend request
	backend_res, backend_err := s.postRequest("/services/haproxy/configuration/backends", add_backend_request_query_params, bytes.NewReader(add_backend_request_body_bytes))
	if backend_err != nil || !isValidStatusCode(backend_res.StatusCode) {
		return errors.New("failed to add backend")
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
		return errors.New("failed to marshal add_server_template_request_body")
	}

	server_template_res, server_template_err := s.postRequest("/services/haproxy/configuration/server_templates", add_server_template_request_query_params, bytes.NewReader(add_server_template_request_body_bytes))
	if server_template_err != nil || !isValidStatusCode(server_template_res.StatusCode) {
		return errors.New("failed to add server template")
	}
	defer server_template_res.Body.Close()
	return nil
}

// Delete Backend from HAProxy configuration
func (s Manager) DeleteBackend(transaction_id string, service_name string, port int) error {
	backend_name := "be_" + service_name + "_" + strconv.Itoa(port)
	// Build query parameterss
	add_backend_request_query_params := QueryParameters{}
	add_backend_request_query_params.add("transaction_id", transaction_id)
	// Delete backend request
	backend_res, backend_err := s.deleteRequest("/services/haproxy/configuration/backends/"+backend_name, add_backend_request_query_params)
	if backend_err != nil || !isValidStatusCode(backend_res.StatusCode) {
		return errors.New("failed to delete backend")
	}
	defer backend_res.Body.Close()
	return nil
}

// Update Backend from HAProxy configuration
func (s Manager) UpdateBackend(transaction_id string,
	old_service_name string, old_port int, old_replicas int,
	new_service_name string, new_port int, new_replicas int,
) error {
	// Delete old backend
	err := s.DeleteBackend(transaction_id, old_service_name, old_port)
	if err != nil {
		return errors.New("failed to delete old backend")
	}
	// Add new backend
	err = s.AddBackend(transaction_id, new_service_name, new_port, new_replicas)
	if err != nil {
		return errors.New("failed to add new backend")
	}

	return nil
}
