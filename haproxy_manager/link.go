package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Add HTTP Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPLink(transaction_id string, backend_name string, domain_name string) error {
	frontend_name := "fe_http"
	// Build query parameterss
	add_backend_switch_request_query_params := QueryParameters{}
	add_backend_switch_request_query_params.add("transaction_id", transaction_id)
	add_backend_switch_request_query_params.add("frontend", frontend_name)
	// Add backend switch request body
	add_backend_switch_request_body := map[string]interface{}{
		"cond":      "if",
		"cond_test": `{ hdr(host) -i ` + domain_name + ` }`,
		"index":     1,
		"name":      backend_name,
	}
	add_backend_switch_request_body_bytes, err := json.Marshal(add_backend_switch_request_body)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send add backend switch request
	backend_switch_res, backend_switch_err := s.postRequest("/services/haproxy/configuration/backend_switching_rules", add_backend_switch_request_query_params, bytes.NewReader(add_backend_switch_request_body_bytes))
	if backend_switch_err != nil || !isValidStatusCode(backend_switch_res.StatusCode) {
		return errors.New("failed to add backend switch")
	}
	defer backend_switch_res.Body.Close()
	return nil
}

// Delete HTTP Link from HAProxy configuration
func (s Manager) DeleteHTTPLink(transaction_id string, backend_name string, domain_name string) error {
	frontend_name := "fe_http"
	// Build query parameterss
	fetch_backend_switch_request_query_params := QueryParameters{}
	fetch_backend_switch_request_query_params.add("transaction_id", transaction_id)
	fetch_backend_switch_request_query_params.add("frontend", frontend_name)
	// Fetch backend switch
	backend_switch_res, backend_switch_err := s.getRequest("/services/haproxy/configuration/backend_switching_rules", fetch_backend_switch_request_query_params)
	if backend_switch_err != nil || !isValidStatusCode(backend_switch_res.StatusCode) {
		return errors.New("failed to fetch backend switch")
	}
	defer backend_switch_res.Body.Close()
	// Parse backend switch
	backend_switch_data := map[string]interface{}{}
	bodyBytes, err := io.ReadAll(backend_switch_res.Body)
	if err != nil {
		return errors.New("failed to read backend switch response body")
	}
	err = json.Unmarshal(bodyBytes, &backend_switch_data)
	if err != nil {
		return errors.New("failed to parse backend switch response body")
	}
	// Find backend switch
	backend_switch_data_array := backend_switch_data["data"].([]interface{})
	backend_switch_data_array_index := -1
	for i, backend_switch_data_array_item := range backend_switch_data_array {
		backend_switch_data_array_item_map := backend_switch_data_array_item.(map[string]interface{})
		if backend_switch_data_array_item_map["name"] == backend_name && strings.Contains(backend_switch_data_array_item_map["cond_test"].(string), domain_name) {
			backend_switch_data_array_index = i
			break
		}
	}
	if backend_switch_data_array_index == -1 {
		return errors.New("failed to find backend switch")
	}
	// Build query parameterss
	delete_backend_switch_request_query_params := QueryParameters{}
	delete_backend_switch_request_query_params.add("transaction_id", transaction_id)
	delete_backend_switch_request_query_params.add("frontend", frontend_name)

	// Delete backend switch
	delete_backend_switch_res, delete_backend_switch_err := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backend_switch_data_array_index), delete_backend_switch_request_query_params)
	if delete_backend_switch_err != nil || !isValidStatusCode(delete_backend_switch_res.StatusCode) {
		return errors.New("failed to delete backend switch")
	}
	defer delete_backend_switch_res.Body.Close()
	return nil
}

// Add HTTPS Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPSLink(transaction_id string, backend_name string, domain_name string) error {
	frontend_name := "fe_https"
	// Build query parameters
	add_backend_switch_request_query_params := QueryParameters{}
	add_backend_switch_request_query_params.add("transaction_id", transaction_id)
	add_backend_switch_request_query_params.add("frontend", frontend_name)
	// Add backend switch request body
	add_backend_switch_request_body := map[string]interface{}{
		"cond":      "if",
		"cond_test": `{ hdr(host) -i ` + domain_name + ` }`,
		"index":     0,
		"name":      backend_name,
	}
	add_backend_switch_request_body_bytes, err := json.Marshal(add_backend_switch_request_body)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send add backend switch request
	backend_switch_res, backend_switch_err := s.postRequest("/services/haproxy/configuration/backend_switching_rules", add_backend_switch_request_query_params, bytes.NewReader(add_backend_switch_request_body_bytes))
	if backend_switch_err != nil || !isValidStatusCode(backend_switch_res.StatusCode) {
		return errors.New("failed to add backend switch")
	}
	defer backend_switch_res.Body.Close()
	return nil
}

// Delete HTTPS Link from HAProxy configuration
func (s Manager) DeleteHTTPSLink(transaction_id string, backend_name string, domain_name string) error {
	frontend_name := "fe_https"
	// Build query parameterss
	fetch_backend_switch_request_query_params := QueryParameters{}
	fetch_backend_switch_request_query_params.add("transaction_id", transaction_id)
	fetch_backend_switch_request_query_params.add("frontend", frontend_name)
	// Fetch backend switch
	backend_switch_res, backend_switch_err := s.getRequest("/services/haproxy/configuration/backend_switching_rules", fetch_backend_switch_request_query_params)
	if backend_switch_err != nil || !isValidStatusCode(backend_switch_res.StatusCode) {
		return errors.New("failed to fetch backend switch")
	}
	defer backend_switch_res.Body.Close()
	// Parse backend switch
	backend_switch_data := map[string]interface{}{}
	bodyBytes, err := io.ReadAll(backend_switch_res.Body)
	if err != nil {
		return errors.New("failed to read backend switch response body")
	}
	err = json.Unmarshal(bodyBytes, &backend_switch_data)
	if err != nil {
		return errors.New("failed to parse backend switch response body")
	}
	// Find backend switch
	backend_switch_data_array := backend_switch_data["data"].([]interface{})
	backend_switch_data_array_index := -1
	for i, backend_switch_data_array_item := range backend_switch_data_array {
		backend_switch_data_array_item_map := backend_switch_data_array_item.(map[string]interface{})
		if backend_switch_data_array_item_map["name"] == backend_name && strings.Contains(backend_switch_data_array_item_map["cond_test"].(string), domain_name) {
			backend_switch_data_array_index = i
			break
		}
	}
	if backend_switch_data_array_index == -1 {
		return errors.New("failed to find backend switch")
	}
	// Build query parameterss
	delete_backend_switch_request_query_params := QueryParameters{}
	delete_backend_switch_request_query_params.add("transaction_id", transaction_id)
	delete_backend_switch_request_query_params.add("frontend", frontend_name)

	// Delete backend switch
	delete_backend_switch_res, delete_backend_switch_err := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(backend_switch_data_array_index), delete_backend_switch_request_query_params)
	if delete_backend_switch_err != nil || !isValidStatusCode(delete_backend_switch_res.StatusCode) {
		return errors.New("failed to delete backend switch")
	}
	defer delete_backend_switch_res.Body.Close()
	return nil
}

// Add TCP Frontend to HAProxy configuration
// -- Manage ACLs with frontend [port{required} and domain_name{optional}]
// -- Manage rules with frontend and backend switch
func (s Manager) AddTCPLink(transaction_id string, backend_name string, port int, domain_name string, listenerMode ListenerMode) error {
	if isPortRestrictedForManualConfig(port) {
		return errors.New("port is restricted for manual configuration")
	}
	frontend_name := ""
	if domain_name == "" {
		frontend_name = "fe_tcp_" + strconv.Itoa(port)
	} else {
		frontend_name = "fe_tcp_" + strconv.Itoa(port) + "_" + domain_name
	}
	// Add TCP Frontend
	add_tcp_frontend_request_query_params := QueryParameters{}
	add_tcp_frontend_request_query_params.add("transaction_id", transaction_id)
	add_tcp_frontend_request_body := map[string]interface{}{
		"maxconn": 2000,
		"mode":    listenerMode,
		"name":    frontend_name,
	}
	if strings.TrimSpace(domain_name) == "" {
		add_tcp_frontend_request_body["default_backend"] = backend_name
	}
	// Create request bytes
	add_tcp_frontend_request_body_bytes, err := json.Marshal(add_tcp_frontend_request_body)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send request
	add_tcp_frontend_res, add_tcp_frontend_err := s.postRequest("/services/haproxy/configuration/frontends", add_tcp_frontend_request_query_params, bytes.NewReader(add_tcp_frontend_request_body_bytes))
	if add_tcp_frontend_err != nil || !isValidStatusCode(add_tcp_frontend_res.StatusCode) {
		return errors.New("failed to add tcp frontend")
	}
	defer add_tcp_frontend_res.Body.Close()

	// Add Port binding
	add_port_binding_request_query_params := QueryParameters{}
	add_port_binding_request_query_params.add("transaction_id", transaction_id)
	add_port_binding_request_query_params.add("frontend", frontend_name)

	add_port_binding_request_body := map[string]interface{}{
		"ssl":  false,
		"port": port,
	}
	// Create request bytes
	add_port_binding_request_body_bytes, err := json.Marshal(add_port_binding_request_body)
	if err != nil {
		return errors.New("failed to marshal add_port_binding_request_body")
	}
	// Send request
	add_port_binding_res, add_port_binding_err := s.postRequest("/services/haproxy/configuration/binds", add_port_binding_request_query_params, bytes.NewReader(add_port_binding_request_body_bytes))
	if add_port_binding_err != nil || !isValidStatusCode(add_port_binding_res.StatusCode) {
		return errors.New("failed to add port binding")
	}
	defer add_port_binding_res.Body.Close()

	if strings.TrimSpace(domain_name) != "" {
		/// Add Backend Switch
		// Build query parameterss
		add_backend_switch_request_query_params := QueryParameters{}
		add_backend_switch_request_query_params.add("transaction_id", transaction_id)
		add_backend_switch_request_query_params.add("frontend", frontend_name)

		// Add backend switch request body
		add_backend_switch_request_body := map[string]interface{}{
			"cond":      "if",
			"cond_test": `{ hdr(host) -i ` + strings.TrimSpace(domain_name) + `:` + strconv.Itoa(port) + ` }`,
			"index":     0,
			"name":      backend_name,
		}
		add_backend_switch_request_body_bytes, err := json.Marshal(add_backend_switch_request_body)
		if err != nil {
			return errors.New("failed to marshal add_backend_switch_request_body")
		}
		// Send add backend switch request
		backend_switch_res, backend_switch_err := s.postRequest("/services/haproxy/configuration/backend_switching_rules", add_backend_switch_request_query_params, bytes.NewReader(add_backend_switch_request_body_bytes))
		if backend_switch_err != nil || !isValidStatusCode(backend_switch_res.StatusCode) {

			return errors.New("failed to add backend switch")
		}
		defer backend_switch_res.Body.Close()
	}

	return nil
}

// Delete TCP Frontend from HAProxy configuration
func (s Manager) DeleteTCPLink(transaction_id string, backend_name string, port int, domain_name string) error {
	if isPortRestrictedForManualConfig(port) {
		return errors.New("port is restricted for manual configuration")
	}
	frontend_name := ""
	if domain_name == "" {
		frontend_name = "fe_tcp_" + strconv.Itoa(port)
	} else {
		frontend_name = "fe_tcp_" + strconv.Itoa(port) + "_" + domain_name
	}
	// Delete TCP Frontend
	delete_tcp_frontend_request_query_params := QueryParameters{}
	delete_tcp_frontend_request_query_params.add("transaction_id", transaction_id)
	delete_tcp_frontend_request_query_params.add("frontend", frontend_name)
	// Send request
	delete_tcp_frontend_res, delete_tcp_frontend_err := s.deleteRequest("/services/haproxy/configuration/frontends/"+frontend_name, delete_tcp_frontend_request_query_params)
	if delete_tcp_frontend_err != nil || !isValidStatusCode(delete_tcp_frontend_res.StatusCode) {
		return errors.New("failed to delete tcp frontend")
	}
	return nil
}
