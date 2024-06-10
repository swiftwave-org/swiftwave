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

func (s Manager) EnableHTTPSRedirection(transactionId string, domainName string) error {
	frontendName := s.GenerateFrontendName(HTTPMode, 80)
	// check if http-request rule already exists
	index, err := s.FetchIndexOfHTTPSRedirection(transactionId, domainName)
	if err != nil {
		return err
	}
	if index != -1 {
		log.Println("HTTPS redirection already enabled")
		return nil
	}
	// build request body
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	reqBody := map[string]interface{}{
		"type":        "redirect",
		"redir_code":  301,
		"redir_type":  "scheme",
		"redir_value": "https",
		"index":       0,
		"cond":        "if",
		"cond_test":   `{ hdr(host) -i ` + strings.TrimSpace(domainName) + ` }`,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	// Send request
	addHTTPSRedirectionRes, addHTTPSRedirectionErr := s.postRequest("/services/haproxy/configuration/http_request_rules", params, bytes.NewReader(reqBodyBytes))
	if addHTTPSRedirectionErr != nil || !isValidStatusCode(addHTTPSRedirectionRes.StatusCode) {
		return errors.New("failed to enable HTTPS redirection")
	}

	return nil
}

func (s Manager) FetchIndexOfHTTPSRedirection(transactionId string, domainName string) (int, error) {
	frontendName := s.GenerateFrontendName(HTTPMode, 80)

	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	// fetch all http-request rules
	httpRequestRulesRes, httpRequestRulesErr := s.getRequest("/services/haproxy/configuration/http_request_rules", params)
	if httpRequestRulesErr != nil || !isValidStatusCode(httpRequestRulesRes.StatusCode) {
		return -1, errors.New("failed to fetch http-request rules")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpRequestRulesRes.Body)
	// parse response
	var httpRequestRulesData map[string]interface{}
	err := json.NewDecoder(httpRequestRulesRes.Body).Decode(&httpRequestRulesData)
	if err != nil {
		return -1, err
	}
	httpRequestRules := httpRequestRulesData["data"].([]interface{})
	condTest := `{ hdr(host) -i ` + strings.TrimSpace(domainName) + ` }`
	// check if http-request rule already exists
	for _, r := range httpRequestRules {
		rule := r.(map[string]interface{})
		if strings.Compare(rule["cond"].(string), "if") == 0 &&
			strings.Compare(rule["cond_test"].(string), condTest) == 0 &&
			strings.Compare(rule["type"].(string), "redirect") == 0 &&
			strings.Compare(rule["redir_type"].(string), "scheme") == 0 &&
			strings.Compare(rule["redir_value"].(string), "https") == 0 {
			return int(rule["index"].(float64)), nil
		}
	}
	return -1, nil

}

func (s Manager) DisableHTTPSRedirection(transactionId string, domainName string) error {
	frontendName := s.GenerateFrontendName(HTTPMode, 80)
	// check if http-request rule already exists
	index, err := s.FetchIndexOfHTTPSRedirection(transactionId, domainName)
	if err != nil {
		return err
	}
	if index == -1 {
		return nil
	}
	// send request
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	disableHTTPSRedirectionRes, disableHTTPSRedirectionErr := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(index), params)
	if disableHTTPSRedirectionErr != nil || !isValidStatusCode(disableHTTPSRedirectionRes.StatusCode) {
		return errors.New("failed to disable HTTPS redirection")
	}
	return nil
}
