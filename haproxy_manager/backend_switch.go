package haproxymanager

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

func (s Manager) AddBackendSwitch(transactionId string, listenerMode ListenerMode, bindPort int, backendName string, domainName string) error {
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", GenerateFrontendName(listenerMode, bindPort))
	var reqBody map[string]interface{}
	// for tcp mode, just add use_backend rule without ACL
	if listenerMode == TCPMode {
		reqBody = map[string]interface{}{
			"index": 0,
			"name":  backendName,
		}
	} else {
		reqBody = map[string]interface{}{
			"cond":      "if",
			"cond_test": `{ hdr(host) -i ` + strings.TrimSpace(domainName) + `:` + strconv.Itoa(bindPort) + ` }`,
			"index":     0,
			"name":      backendName,
		}
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	// Send request
	addBackendSwitchRes, addBackendSwitchErr := s.postRequest("/services/haproxy/configuration/backend_switching_rules", params, bytes.NewReader(reqBodyBytes))
	if addBackendSwitchErr != nil || !isValidStatusCode(addBackendSwitchRes.StatusCode) {
		return addBackendSwitchErr
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(addBackendSwitchRes.Body)
	return nil
}

func (s Manager) FetchBackendSwitchIndex(transactionId string, listenerMode ListenerMode, bindPort int, backendName string, domainName string) (int, error) {
	return s.FetchBackendSwitchIndexByName(transactionId, GenerateFrontendName(listenerMode, bindPort), bindPort, backendName, domainName)
}

func (s Manager) FetchBackendSwitchIndexByName(transactionId string, frontendName string, bindPort int, backendName string, domainName string) (int, error) {
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", frontendName)
	// Send request
	getBackendSwitchRes, getBackendSwitchErr := s.getRequest("/services/haproxy/configuration/backend_switching_rules", params)
	if getBackendSwitchErr != nil || !isValidStatusCode(getBackendSwitchRes.StatusCode) {
		return -1, getBackendSwitchErr
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(getBackendSwitchRes.Body)
	// Parse response
	var backendSwitchRulesData map[string]interface{}
	err := json.NewDecoder(getBackendSwitchRes.Body).Decode(&backendSwitchRulesData)
	if err != nil {
		return -1, err
	}
	backendSwitchRules := backendSwitchRulesData["data"].([]interface{})
	condTest := `{ hdr(host) -i ` + strings.TrimSpace(domainName) + `:` + strconv.Itoa(bindPort) + ` }`
	if bindPort == 80 || bindPort == 443 {
		condTest = `{ hdr(host) -i ` + strings.TrimSpace(domainName) + ` }`
	}
	for _, r := range backendSwitchRules {
		rule := r.(map[string]interface{})
		if rule["name"] == backendName &&
			rule["cond"] == "if" &&
			rule["cond_test"] == condTest {
			return int(rule["index"].(float64)), nil
		}
	}
	return -1, nil
}

func (s Manager) DeleteBackendSwitch(transactionId string, listenerMode ListenerMode, bindPort int, backendName string, domainName string) error {
	// check if frontend already exists
	isFrontendExist, _ := s.IsFrontendExist(transactionId, listenerMode, bindPort)
	if !isFrontendExist {
		return nil
	}
	// fetch backend switch index
	index, err := s.FetchBackendSwitchIndex(transactionId, listenerMode, bindPort, backendName, domainName)
	if err != nil {
		return err
	}
	// not found
	if index == -1 {
		return nil
	}
	// Build query parameters
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", GenerateFrontendName(listenerMode, bindPort))
	deleteReq, err := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(index), params)
	if err != nil || !isValidStatusCode(deleteReq.StatusCode) {
		return err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(deleteReq.Body)
	return nil
}
