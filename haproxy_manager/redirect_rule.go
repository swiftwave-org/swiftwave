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
