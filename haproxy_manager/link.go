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

// AddTCPLink Add TCP Frontend to HAProxy configuration
// -- Manage ACLs with frontend [port{required} and domain_name{optional}]
// -- Manage rules with frontend and backend switch
func (s Manager) AddTCPLink(transactionId string, backendName string, port int, domainName string, listenerMode ListenerMode, restrictedPorts []int) error {
	if IsPortRestrictedForManualConfig(port, restrictedPorts) {
		return errors.New("port is restricted for manual configuration")
	}
	frontendName := ""
	if domainName == "" {
		frontendName = "fe_tcp_" + strconv.Itoa(port)
	} else {
		frontendName = "fe_tcp_" + strconv.Itoa(port) + "_" + domainName
	}
	// Add TCP Frontend
	addTcpFrontendRequestQueryParams := QueryParameters{}
	addTcpFrontendRequestQueryParams.add("transaction_id", transactionId)
	addTcpFrontendRequestBody := map[string]interface{}{
		"maxconn": 2000,
		"mode":    listenerMode,
		"name":    frontendName,
	}
	if strings.TrimSpace(domainName) == "" {
		addTcpFrontendRequestBody["default_backend"] = backendName
	}
	// Create request bytes
	addTcpFrontendRequestBodyBytes, err := json.Marshal(addTcpFrontendRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_backend_switch_request_body")
	}
	// Send request
	addTcpFrontendRes, addTcpFrontendErr := s.postRequest("/services/haproxy/configuration/frontends", addTcpFrontendRequestQueryParams, bytes.NewReader(addTcpFrontendRequestBodyBytes))
	if addTcpFrontendErr != nil || !isValidStatusCode(addTcpFrontendRes.StatusCode) {
		return errors.New("failed to add tcp frontend")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddTCPLink: ", err)
		}
	}(addTcpFrontendRes.Body)

	// Add Port binding
	addPortBindingRequestQueryParams := QueryParameters{}
	addPortBindingRequestQueryParams.add("transaction_id", transactionId)
	addPortBindingRequestQueryParams.add("frontend", frontendName)

	addPortBindingRequestBody := map[string]interface{}{
		"ssl":  false,
		"port": port,
	}
	// Create request bytes
	addPortBindingRequestBodyBytes, err := json.Marshal(addPortBindingRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_port_binding_request_body")
	}
	// Send request
	addPortBindingRes, addPortBindingErr := s.postRequest("/services/haproxy/configuration/binds", addPortBindingRequestQueryParams, bytes.NewReader(addPortBindingRequestBodyBytes))
	if addPortBindingErr != nil || !isValidStatusCode(addPortBindingRes.StatusCode) {
		return errors.New("failed to add port binding")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddTCPLink: ", err)
		}
	}(addPortBindingRes.Body)

	if strings.TrimSpace(domainName) != "" {
		/// Add Backend Switch
		// Build query parameters
		addBackendSwitchRequestQueryParams := QueryParameters{}
		addBackendSwitchRequestQueryParams.add("transaction_id", transactionId)
		addBackendSwitchRequestQueryParams.add("frontend", frontendName)

		// Add backend switch request body
		addBackendSwitchRequestBody := map[string]interface{}{
			"cond":      "if",
			"cond_test": `{ hdr(host) -i ` + strings.TrimSpace(domainName) + `:` + strconv.Itoa(port) + ` }`,
			"index":     0,
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
				log.Println("[haproxy_manager] AddTCPLink: ", err)
			}
		}(backendSwitchRes.Body)
	}

	return nil
}

// DeleteTCPLink Delete TCP Frontend from HAProxy configuration
func (s Manager) DeleteTCPLink(transactionId string, backendName string, port int, domainName string, restrictedPorts []int) error {
	_ = backendName
	if IsPortRestrictedForManualConfig(port, restrictedPorts) {
		return errors.New("port is restricted for manual configuration")
	}
	frontendName := ""
	if domainName == "" {
		frontendName = "fe_tcp_" + strconv.Itoa(port)
	} else {
		frontendName = "fe_tcp_" + strconv.Itoa(port) + "_" + domainName
	}
	// Delete TCP Frontend
	deleteTcpFrontendRequestQueryParams := QueryParameters{}
	deleteTcpFrontendRequestQueryParams.add("transaction_id", transactionId)
	deleteTcpFrontendRequestQueryParams.add("frontend", frontendName)
	// Send request
	deleteTcpFrontendRes, deleteTcpFrontendErr := s.deleteRequest("/services/haproxy/configuration/frontends/"+frontendName, deleteTcpFrontendRequestQueryParams)
	if deleteTcpFrontendErr != nil {
		return errors.New("failed to delete tcp frontend")
	}
	if deleteTcpFrontendRes.StatusCode == 404 {
		return nil
	} else if !isValidStatusCode(deleteTcpFrontendRes.StatusCode) {
		return errors.New("failed to delete tcp frontend")
	}
	return nil
}

// AddHTTPRedirectRule Add HTTP Redirect Rule
func (s Manager) AddHTTPRedirectRule(transactionId string, matchDomain string, redirectUrl string) error {
	if strings.TrimSpace(matchDomain) == "" {
		return errors.New("match domain is required")
	}
	if strings.TrimSpace(redirectUrl) == "" {
		return errors.New("redirect domain is required")
	}
	// Add HTTP Redirect Rule
	addHttpRedirectRuleRequestQueryParams := QueryParameters{}
	addHttpRedirectRuleRequestQueryParams.add("transaction_id", transactionId)
	addHttpRedirectRuleRequestQueryParams.add("parent_name", "fe_http")
	addHttpRedirectRuleRequestQueryParams.add("parent_type", "frontend")
	addHttpRedirectRuleRequestBody := map[string]interface{}{
		"type":        "redirect",
		"redir_code":  302,
		"redir_type":  "location",
		"redir_value": redirectUrl,
		"index":       0,
		"cond":        "if",
		"cond_test":   `{ hdr(host) -i ` + strings.TrimSpace(matchDomain) + ` } !letsencrypt-acl`,
	}
	// Create request bytes
	addHttpRedirectRuleRequestBodyBytes, err := json.Marshal(addHttpRedirectRuleRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_http_redirect_rule_request_body")
	}
	// Send request
	addHttpRedirectRuleRes, addHttpRedirectRuleErr := s.postRequest("/services/haproxy/configuration/http_request_rules", addHttpRedirectRuleRequestQueryParams, bytes.NewReader(addHttpRedirectRuleRequestBodyBytes))
	if addHttpRedirectRuleErr != nil || !isValidStatusCode(addHttpRedirectRuleRes.StatusCode) {
		return errors.New("failed to add http redirect rule")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddHTTPRedirectRule: ", err)
		}
	}(addHttpRedirectRuleRes.Body)
	return nil
}

// AddHTTPSRedirectRule Add HTTPS Redirect Rule
func (s Manager) AddHTTPSRedirectRule(transactionId string, matchDomain string, redirectUrl string) error {
	if strings.TrimSpace(matchDomain) == "" {
		return errors.New("match domain is required")
	}
	if strings.TrimSpace(redirectUrl) == "" {
		return errors.New("redirect url is required")
	}
	// Add HTTPS Redirect Rule
	addHttpsRedirectRuleRequestQueryParams := QueryParameters{}
	addHttpsRedirectRuleRequestQueryParams.add("transaction_id", transactionId)
	addHttpsRedirectRuleRequestQueryParams.add("parent_name", "fe_https")
	addHttpsRedirectRuleRequestQueryParams.add("parent_type", "frontend")
	addHttpsRedirectRuleRequestBody := map[string]interface{}{
		"type":        "redirect",
		"redir_code":  302,
		"redir_type":  "location",
		"redir_value": redirectUrl,
		"index":       0,
		"cond":        "if",
		"cond_test":   `{ hdr(host) -i ` + strings.TrimSpace(matchDomain) + ` } !letsencrypt-acl`,
	}
	// Create request bytes
	addHttpsRedirectRuleRequestBodyBytes, err := json.Marshal(addHttpsRedirectRuleRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_https_redirect_rule_request_body")
	}
	// Send request
	addHttpsRedirectRuleRes, addHttpsRedirectRuleErr := s.postRequest("/services/haproxy/configuration/http_request_rules", addHttpsRedirectRuleRequestQueryParams, bytes.NewReader(addHttpsRedirectRuleRequestBodyBytes))
	if addHttpsRedirectRuleErr != nil || !isValidStatusCode(addHttpsRedirectRuleRes.StatusCode) {
		return errors.New("failed to add https redirect rule")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] AddHTTPSRedirectRule: ", err)
		}
	}(addHttpsRedirectRuleRes.Body)
	return nil
}

// DeleteHTTPRedirectRule Delete HTTP Redirect Rule
func (s Manager) DeleteHTTPRedirectRule(transactionId string, matchDomain string) error {
	if strings.TrimSpace(matchDomain) == "" {
		return errors.New("match domain is required")
	}
	// Fetch all HTTP Redirect Rules
	getHttpRedirectRulesRequestQueryParams := QueryParameters{}
	getHttpRedirectRulesRequestQueryParams.add("transaction_id", transactionId)
	getHttpRedirectRulesRequestQueryParams.add("parent_name", "fe_http")
	getHttpRedirectRulesRequestQueryParams.add("parent_type", "frontend")
	getHttpRedirectRulesRes, getHttpRedirectRulesErr := s.getRequest("/services/haproxy/configuration/http_request_rules", getHttpRedirectRulesRequestQueryParams)
	if getHttpRedirectRulesErr != nil || !isValidStatusCode(getHttpRedirectRulesRes.StatusCode) {
		return errors.New("failed to fetch http redirect rules")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPRedirectRule: ", err)
		}
	}(getHttpRedirectRulesRes.Body)
	getHttpRedirectRulesResBody, getHttpRedirectRulesResBodyErr := io.ReadAll(getHttpRedirectRulesRes.Body)
	if getHttpRedirectRulesResBodyErr != nil {
		return errors.New("failed to read http redirect rules response body")
	}
	getHttpRedirectRulesResBodyJson := map[string]interface{}{}
	getHttpRedirectRulesResBodyJsonErr := json.Unmarshal(getHttpRedirectRulesResBody, &getHttpRedirectRulesResBodyJson)
	if getHttpRedirectRulesResBodyJsonErr != nil {
		log.Println(getHttpRedirectRulesResBodyJsonErr)
		return errors.New("[haproxy_manager] DeleteHTTPRedirectRule: failed to unmarshal http redirect rules response body")
	}
	// Find index of HTTP Redirect Rule
	index := -1
	getHttpRedirectRules := getHttpRedirectRulesResBodyJson["data"].([]interface{})
	for _, httpRedirectRule := range getHttpRedirectRules {
		httpRedirectRuleItem := httpRedirectRule.(map[string]interface{})
		if httpRedirectRuleItem["cond_test"] == `{ hdr(host) -i `+strings.TrimSpace(matchDomain)+` } !letsencrypt-acl` {
			index = int(httpRedirectRuleItem["index"].(float64))
			break
		}
	}
	// Delete HTTP Redirect Rule
	if index != -1 {
		deleteHttpRedirectRuleRequestQueryParams := QueryParameters{}
		deleteHttpRedirectRuleRequestQueryParams.add("transaction_id", transactionId)
		deleteHttpRedirectRuleRequestQueryParams.add("parent_name", "fe_http")
		deleteHttpRedirectRuleRequestQueryParams.add("parent_type", "frontend")
		// Send request
		deleteHttpRedirectRuleRes, deleteHttpRedirectRuleErr := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(index), deleteHttpRedirectRuleRequestQueryParams)
		if deleteHttpRedirectRuleErr != nil || !isValidStatusCode(deleteHttpRedirectRuleRes.StatusCode) {
			return errors.New("failed to delete http redirect rule")
		}
		return nil
	}
	return nil
}

// DeleteHTTPSRedirectRule Delete HTTPS Redirect Rule
func (s Manager) DeleteHTTPSRedirectRule(transactionId string, matchDomain string) error {
	if strings.TrimSpace(matchDomain) == "" {
		return errors.New("match domain is required")
	}
	// Fetch all HTTPS Redirect Rules
	getHttpsRedirectRulesRequestQueryParams := QueryParameters{}
	getHttpsRedirectRulesRequestQueryParams.add("transaction_id", transactionId)
	getHttpsRedirectRulesRequestQueryParams.add("parent_name", "fe_https")
	getHttpsRedirectRulesRequestQueryParams.add("parent_type", "frontend")
	getHttpsRedirectRulesRes, getHttpsRedirectRulesErr := s.getRequest("/services/haproxy/configuration/http_request_rules", getHttpsRedirectRulesRequestQueryParams)
	if getHttpsRedirectRulesErr != nil || !isValidStatusCode(getHttpsRedirectRulesRes.StatusCode) {
		return errors.New("failed to fetch https redirect rules")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteHTTPSRedirectRule: ", err)
		}
	}(getHttpsRedirectRulesRes.Body)
	getHttpsRedirectRulesResBody, getHttpsRedirectRulesResBodyErr := io.ReadAll(getHttpsRedirectRulesRes.Body)
	if getHttpsRedirectRulesResBodyErr != nil {
		return errors.New("failed to read https redirect rules response body")
	}
	getHttpsRedirectRulesResBodyJson := map[string]interface{}{}
	getHttpsRedirectRulesResBodyJsonErr := json.Unmarshal(getHttpsRedirectRulesResBody, &getHttpsRedirectRulesResBodyJson)
	if getHttpsRedirectRulesResBodyJsonErr != nil {
		log.Println(getHttpsRedirectRulesResBodyJsonErr)
		return errors.New("failed to unmarshal https redirect rules response body")
	}
	// Find index of HTTPS Redirect Rule
	index := -1
	getHttpsRedirectRules := getHttpsRedirectRulesResBodyJson["data"].([]interface{})
	for _, httpsRedirectRule := range getHttpsRedirectRules {
		httpsRedirectRuleItem := httpsRedirectRule.(map[string]interface{})
		if httpsRedirectRuleItem["cond_test"] == `{ hdr(host) -i `+strings.TrimSpace(matchDomain)+` } !letsencrypt-acl` {
			index = int(httpsRedirectRuleItem["index"].(float64))
			break
		}
	}
	// Delete HTTPS Redirect Rule
	if index != -1 {
		deleteHttpsRedirectRuleRequestQueryParams := QueryParameters{}
		deleteHttpsRedirectRuleRequestQueryParams.add("transaction_id", transactionId)
		deleteHttpsRedirectRuleRequestQueryParams.add("parent_name", "fe_https")
		deleteHttpsRedirectRuleRequestQueryParams.add("parent_type", "frontend")
		// Send request
		deleteHttpsRedirectRuleRes, deleteHttpsRedirectRule := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(index), deleteHttpsRedirectRuleRequestQueryParams)
		if deleteHttpsRedirectRule != nil || !isValidStatusCode(deleteHttpsRedirectRuleRes.StatusCode) {
			return errors.New("failed to delete https redirect rule")
		}
		return nil
	}
	return nil
}
