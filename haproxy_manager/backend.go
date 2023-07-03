package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Add Backend to HAProxy configuration
// -- Manage server template with backend
func (s HAProxySocket) AddBackend(transaction_id string, service_name string, port int, replicas int) error {
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
	// Add server template query parameters
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
func (s HAProxySocket) DeleteBackend(transaction_id string, service_name string, port int) error {
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
func (s HAProxySocket) UpdateBackend(transaction_id string,
	old_service_name string, old_port int, old_replicas int,
	new_service_name string, new_port int, new_replicas int,
) error {
	// Backend name
	old_backend_name := "be_" + old_service_name + "_" + strconv.Itoa(old_port)
	new_backend_name := "be_" + new_service_name + "_" + strconv.Itoa(new_port)
	// Check if backend update required

	// #TODO fix this
	if old_backend_name != new_backend_name {
		// Build query parameterss
		update_backend_request_query_params := QueryParameters{}
		update_backend_request_query_params.add("transaction_id", transaction_id)
		// Add backend request body
		update_backend_request_body := map[string]interface{}{
			"name": new_backend_name,
			"balance": map[string]interface{}{
				"algorithm": "roundrobin",
			},
		}
		update_backend_request_body_bytes, err := json.Marshal(update_backend_request_body)
		if err != nil {
			return errors.New("failed to marshal add_backend_request_body")
		}
		// Send add backend request
		backend_res, backend_err := s.putRequest("/services/haproxy/configuration/backends/"+old_backend_name, update_backend_request_query_params, bytes.NewReader(update_backend_request_body_bytes))
		if backend_err != nil || !isValidStatusCode(backend_res.StatusCode) {
			return errors.New("failed to update backend")
		}
		defer backend_res.Body.Close()
		// print body
		body, err := io.ReadAll(backend_res.Body)
		if err != nil {
			return errors.New("failed to read body")
		}
		fmt.Println("hemlo")
		fmt.Println(string(body))
	}

	// Check if server template update required
	if old_service_name != new_service_name || old_port != new_port || old_replicas != new_replicas {
		// Server template prefix
		old_server_template_prefix := old_service_name + "_container-"
		new_server_template_prefix := new_service_name + "_container-"

		// Update server template query parameters
		update_server_template_request_query_params := QueryParameters{}
		update_server_template_request_query_params.add("transaction_id", transaction_id)
		update_server_template_request_query_params.add("backend", old_backend_name)
		if new_replicas <= 0 {
			new_replicas = 1
		}
		// Update server template request body
		update_server_template_request_body := map[string]interface{}{
			"prefix":       new_server_template_prefix,
			"fqdn":         new_service_name,
			"port":         new_port,
			"check":        "disabled",
			"resolvers":    "docker",
			"init-addr":    "libc,none",
			"num_or_range": strconv.Itoa(new_replicas),
		}
		update_server_template_request_body_bytes, err := json.Marshal(update_server_template_request_body)
		if err != nil {
			return errors.New("failed to marshal update_server_template_request_body")
		}

		server_template_res, server_template_err := s.putRequest("/services/haproxy/configuration/server_templates/"+old_server_template_prefix, update_server_template_request_query_params, bytes.NewReader(update_server_template_request_body_bytes))
		if server_template_err != nil || !isValidStatusCode(server_template_res.StatusCode) {
			fmt.Println(server_template_res.StatusCode)
			fmt.Println(old_server_template_prefix)
			fmt.Println(server_template_err)
			return errors.New("failed to update server template")
		}
		defer server_template_res.Body.Close()
	}

	return nil
}
