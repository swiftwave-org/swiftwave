package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

func (s Manager) AddBackendSwitch(transactionId string, listenerMode ListenerMode, bindPort int, backendName string, domainName string) error {
	// check if backend switch already exists
	index, err := s.FetchBackendSwitchIndex(transactionId, listenerMode, bindPort, backendName, domainName)
	if err != nil {
		return err
	}
	if index != -1 {
		return nil
	}
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", s.GenerateFrontendName(listenerMode, bindPort))
	var reqBody map[string]interface{}
	// for tcp mode, just add use_backend rule without ACL
	if listenerMode == TCPMode {
		reqBody = map[string]interface{}{
			"index": 0,
			"name":  backendName,
		}
	} else {
		condTest := `{ hdr(host) -i ` + strings.TrimSpace(domainName) + `:` + strconv.Itoa(bindPort) + ` }`
		if bindPort == 80 || bindPort == 443 {
			condTest = `{ hdr(host) -i ` + strings.TrimSpace(domainName) + ` }`
		}
		aclIndex := 0
		if listenerMode == HTTPMode && (bindPort == 80 || bindPort == 443) {
			aclIndex = 1
		}
		reqBody = map[string]interface{}{
			"cond":      "if",
			"cond_test": condTest,
			"index":     aclIndex,
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
		return errors.New("failed to add backend switch")
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(addBackendSwitchRes.Body)
	return nil
}

func (s Manager) FetchBackendSwitchIndex(transactionId string, listenerMode ListenerMode, bindPort int, backendName string, domainName string) (int, error) {
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", frontendName)
	// Send request
	getBackendSwitchRes, getBackendSwitchErr := s.getRequest("/services/haproxy/configuration/backend_switching_rules", params)
	if getBackendSwitchErr != nil || !isValidStatusCode(getBackendSwitchRes.StatusCode) {
		return -1, errors.New("failed to fetch backend switch index")
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

		if listenerMode == HTTPMode {
			if rule["name"] == backendName &&
				rule["cond"] == "if" &&
				rule["cond_test"] == condTest {
				return int(rule["index"].(float64)), nil
			}
		} else {
			if rule["name"] == backendName {
				return int(rule["index"].(float64)), nil
			}
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
	params.add("frontend", s.GenerateFrontendName(listenerMode, bindPort))
	deleteReq, err := s.deleteRequest("/services/haproxy/configuration/backend_switching_rules/"+strconv.Itoa(index), params)
	if err != nil || !isValidStatusCode(deleteReq.StatusCode) {
		return errors.New("failed to delete backend switch")
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(deleteReq.Body)
	return nil
}
