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
func (s Manager) AddHTTPLink(transaction_id string, backend_name string, domain_name string) error {
	frontend_name := "fe_http"
	// Build query parameters
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

// DeleteHTTPLink Delete HTTP Link from HAProxy configuration
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

// AddHTTPSLink Add HTTPS Link [Backend Switch] to HAProxy configuration
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

// DeleteHTTPSLink Delete HTTPS Link from HAProxy configuration
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

// AddTCPLink Add TCP Frontend to HAProxy configuration
// -- Manage ACLs with frontend [port{required} and domain_name{optional}]
// -- Manage rules with frontend and backend switch
func (s Manager) AddTCPLink(transaction_id string, backend_name string, port int, domain_name string, listenerMode ListenerMode, restrictedPorts []int) error {
	if IsPortRestrictedForManualConfig(port, restrictedPorts) {
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
func (s Manager) DeleteTCPLink(transaction_id string, backend_name string, port int, domain_name string, restrictedPorts []int) error {
	if IsPortRestrictedForManualConfig(port, restrictedPorts) {
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
	if delete_tcp_frontend_err != nil {
		return errors.New("failed to delete tcp frontend")
	}
	if delete_tcp_frontend_res.StatusCode == 404 {
		return nil
	} else if !isValidStatusCode(delete_tcp_frontend_res.StatusCode) {
		return errors.New("failed to delete tcp frontend")
	}
	return nil
}

// Add HTTP Redirect Rule
func (s Manager) AddHTTPRedirectRule(transaction_id string, match_domain string, redirect_url string) error {
	if strings.TrimSpace(match_domain) == "" {
		return errors.New("match domain is required")
	}
	if strings.TrimSpace(redirect_url) == "" {
		return errors.New("redirect domain is required")
	}
	// Add HTTP Redirect Rule
	add_http_redirect_rule_request_query_params := QueryParameters{}
	add_http_redirect_rule_request_query_params.add("transaction_id", transaction_id)
	add_http_redirect_rule_request_query_params.add("parent_name", "fe_http")
	add_http_redirect_rule_request_query_params.add("parent_type", "frontend")
	add_http_redirect_rule_request_body := map[string]interface{}{
		"type":        "redirect",
		"redir_code":  302,
		"redir_type":  "location",
		"redir_value": redirect_url,
		"index":       0,
		"cond":        "if",
		"cond_test":   `{ hdr(host) -i ` + strings.TrimSpace(match_domain) + ` } !letsencrypt-acl`,
	}
	// Create request bytes
	add_http_redirect_rule_request_body_bytes, err := json.Marshal(add_http_redirect_rule_request_body)
	if err != nil {
		return errors.New("failed to marshal add_http_redirect_rule_request_body")
	}
	// Send request
	add_http_redirect_rule_res, add_http_redirect_rule_err := s.postRequest("/services/haproxy/configuration/http_request_rules", add_http_redirect_rule_request_query_params, bytes.NewReader(add_http_redirect_rule_request_body_bytes))
	if add_http_redirect_rule_err != nil || !isValidStatusCode(add_http_redirect_rule_res.StatusCode) {
		return errors.New("failed to add http redirect rule")
	}
	defer add_http_redirect_rule_res.Body.Close()
	return nil
}

// Add HTTPS Redirect Rule
func (s Manager) AddHTTPSRedirectRule(transaction_id string, match_domain string, redirect_url string) error {
	if strings.TrimSpace(match_domain) == "" {
		return errors.New("match domain is required")
	}
	if strings.TrimSpace(redirect_url) == "" {
		return errors.New("redirect url is required")
	}
	// Add HTTPS Redirect Rule
	add_https_redirect_rule_request_query_params := QueryParameters{}
	add_https_redirect_rule_request_query_params.add("transaction_id", transaction_id)
	add_https_redirect_rule_request_query_params.add("parent_name", "fe_https")
	add_https_redirect_rule_request_query_params.add("parent_type", "frontend")
	add_https_redirect_rule_request_body := map[string]interface{}{
		"type":        "redirect",
		"redir_code":  302,
		"redir_type":  "location",
		"redir_value": redirect_url,
		"index":       0,
		"cond":        "if",
		"cond_test":   `{ hdr(host) -i ` + strings.TrimSpace(match_domain) + ` } !letsencrypt-acl`,
	}
	// Create request bytes
	add_https_redirect_rule_request_body_bytes, err := json.Marshal(add_https_redirect_rule_request_body)
	if err != nil {
		return errors.New("failed to marshal add_https_redirect_rule_request_body")
	}
	// Send request
	add_https_redirect_rule_res, add_https_redirect_rule_err := s.postRequest("/services/haproxy/configuration/http_request_rules", add_https_redirect_rule_request_query_params, bytes.NewReader(add_https_redirect_rule_request_body_bytes))
	if add_https_redirect_rule_err != nil || !isValidStatusCode(add_https_redirect_rule_res.StatusCode) {
		return errors.New("failed to add https redirect rule")
	}
	defer add_https_redirect_rule_res.Body.Close()
	return nil
}

// Delete HTTP Redirect Rule
func (s Manager) DeleteHTTPRedirectRule(transaction_id string, match_domain string) error {
	if strings.TrimSpace(match_domain) == "" {
		return errors.New("match domain is required")
	}
	// Fetch all HTTP Redirect Rules
	get_http_redirect_rules_request_query_params := QueryParameters{}
	get_http_redirect_rules_request_query_params.add("transaction_id", transaction_id)
	get_http_redirect_rules_request_query_params.add("parent_name", "fe_http")
	get_http_redirect_rules_request_query_params.add("parent_type", "frontend")
	get_http_redirect_rules_res, get_http_redirect_rules_err := s.getRequest("/services/haproxy/configuration/http_request_rules", get_http_redirect_rules_request_query_params)
	if get_http_redirect_rules_err != nil || !isValidStatusCode(get_http_redirect_rules_res.StatusCode) {
		return errors.New("failed to fetch http redirect rules")
	}
	defer get_http_redirect_rules_res.Body.Close()
	get_http_redirect_rules_res_body, get_http_redirect_rules_res_body_err := io.ReadAll(get_http_redirect_rules_res.Body)
	if get_http_redirect_rules_res_body_err != nil {
		return errors.New("failed to read http redirect rules response body")
	}
	get_http_redirect_rules_res_body_json := map[string]interface{}{}
	get_http_redirect_rules_res_body_json_err := json.Unmarshal(get_http_redirect_rules_res_body, &get_http_redirect_rules_res_body_json)
	if get_http_redirect_rules_res_body_json_err != nil {
		log.Println(get_http_redirect_rules_res_body_json_err)
		return errors.New("failed to unmarshal http redirect rules response body")
	}
	// Find index of HTTP Redirect Rule
	index := -1
	get_http_redirect_rules := get_http_redirect_rules_res_body_json["data"].([]interface{})
	for _, http_redirect_rule := range get_http_redirect_rules {
		http_redirect_rule_item := http_redirect_rule.(map[string]interface{})
		if http_redirect_rule_item["cond_test"] == `{ hdr(host) -i `+strings.TrimSpace(match_domain)+` } !letsencrypt-acl` {
			index = int(http_redirect_rule_item["index"].(float64))
			break
		}
	}
	// Delete HTTP Redirect Rule
	if index != -1 {
		delete_http_redirect_rule_request_query_params := QueryParameters{}
		delete_http_redirect_rule_request_query_params.add("transaction_id", transaction_id)
		delete_http_redirect_rule_request_query_params.add("parent_name", "fe_http")
		delete_http_redirect_rule_request_query_params.add("parent_type", "frontend")
		// Send request
		delete_http_redirect_rule_res, delete_http_redirect_rule_err := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(index), delete_http_redirect_rule_request_query_params)
		if delete_http_redirect_rule_err != nil || !isValidStatusCode(delete_http_redirect_rule_res.StatusCode) {
			return errors.New("failed to delete http redirect rule")
		}
		return nil
	}
	return nil
}

// Delete HTTPS Redirect Rule
func (s Manager) DeleteHTTPSRedirectRule(transaction_id string, match_domain string) error {
	if strings.TrimSpace(match_domain) == "" {
		return errors.New("match domain is required")
	}
	// Fetch all HTTPS Redirect Rules
	get_https_redirect_rules_request_query_params := QueryParameters{}
	get_https_redirect_rules_request_query_params.add("transaction_id", transaction_id)
	get_https_redirect_rules_request_query_params.add("parent_name", "fe_https")
	get_https_redirect_rules_request_query_params.add("parent_type", "frontend")
	get_https_redirect_rules_res, get_https_redirect_rules_err := s.getRequest("/services/haproxy/configuration/http_request_rules", get_https_redirect_rules_request_query_params)
	if get_https_redirect_rules_err != nil || !isValidStatusCode(get_https_redirect_rules_res.StatusCode) {
		return errors.New("failed to fetch https redirect rules")
	}
	defer get_https_redirect_rules_res.Body.Close()
	get_https_redirect_rules_res_body, get_https_redirect_rules_res_body_err := io.ReadAll(get_https_redirect_rules_res.Body)
	if get_https_redirect_rules_res_body_err != nil {
		return errors.New("failed to read https redirect rules response body")
	}
	get_https_redirect_rules_res_body_json := map[string]interface{}{}
	get_https_redirect_rules_res_body_json_err := json.Unmarshal(get_https_redirect_rules_res_body, &get_https_redirect_rules_res_body_json)
	if get_https_redirect_rules_res_body_json_err != nil {
		log.Println(get_https_redirect_rules_res_body_json_err)
		return errors.New("failed to unmarshal https redirect rules response body")
	}
	// Find index of HTTPS Redirect Rule
	index := -1
	get_https_redirect_rules := get_https_redirect_rules_res_body_json["data"].([]interface{})
	for _, https_redirect_rule := range get_https_redirect_rules {
		https_redirect_rule_item := https_redirect_rule.(map[string]interface{})
		if https_redirect_rule_item["cond_test"] == `{ hdr(host) -i `+strings.TrimSpace(match_domain)+` } !letsencrypt-acl` {
			index = int(https_redirect_rule_item["index"].(float64))
			break
		}
	}
	// Delete HTTPS Redirect Rule
	if index != -1 {
		delete_https_redirect_rule_request_query_params := QueryParameters{}
		delete_https_redirect_rule_request_query_params.add("transaction_id", transaction_id)
		delete_https_redirect_rule_request_query_params.add("parent_name", "fe_https")
		delete_https_redirect_rule_request_query_params.add("parent_type", "frontend")
		// Send request
		delete_https_redirect_rule_res, delete_https_redirect_rule := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(index), delete_https_redirect_rule_request_query_params)
		if delete_https_redirect_rule != nil || !isValidStatusCode(delete_https_redirect_rule_res.StatusCode) {
			return errors.New("failed to delete https redirect rule")
		}
		return nil
	}
	return nil
}
