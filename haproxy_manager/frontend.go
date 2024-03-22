package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

var defaultBackend = "error_backend"

func (s Manager) GenerateFrontendName(listenerMode ListenerMode, port int) string {
	if listenerMode == HTTPMode {
		if port == 80 {
			return "fe_http"
		} else if port == 443 {
			return "fe_https"
		}
	}
	return "fe_" + string(listenerMode) + "_" + strconv.Itoa(port)
}

func (s Manager) AddFrontend(transactionId string, listenerMode ListenerMode, bindPort int, restrictedPorts []int) error {
	if bindPort == 80 || bindPort == 443 {
		if listenerMode == TCPMode {
			return errors.New("frontend with tcp mode cannot be created for port 80 or 443")
		}
		return nil
	}
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	if IsPortRestrictedForManualConfig(bindPort, restrictedPorts) {
		return errors.New("port is restricted for manual configuration")
	}
	// check if frontend already exists
	isFrontendExist, _ := s.IsFrontendExist(transactionId, listenerMode, bindPort)
	if isFrontendExist {
		return nil
	}
	// if frontend with tcp/http mode already exists, then raise error
	if listenerMode == TCPMode {
		isConflictingFrontendExist, err := s.IsFrontendExist(transactionId, HTTPMode, bindPort)
		if err != nil {
			return err
		}
		if isConflictingFrontendExist {
			return errors.New("frontend with http mode already exists")
		}
	} else if listenerMode == HTTPMode {
		isConflictingFrontendExist, err := s.IsFrontendExist(transactionId, TCPMode, bindPort)
		if err != nil {
			return err
		}
		if isConflictingFrontendExist {
			return errors.New("frontend with tcp mode already exists")
		}
	}
	// create frontend
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	body := map[string]interface{}{
		"maxconn": 6000,
		"mode":    listenerMode,
		"name":    frontendName,
	}
	body["default_backend"] = defaultBackend
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return errors.New("failed to marshal frontend create request body")
	}
	// Send request
	addTcpFrontendRes, addTcpFrontendErr := s.postRequest("/services/haproxy/configuration/frontends", params, bytes.NewReader(bodyBytes))
	if addTcpFrontendErr != nil || !isValidStatusCode(addTcpFrontendRes.StatusCode) {
		// 409 status code means that frontend already exists
		if addTcpFrontendRes.StatusCode == 409 {
			return nil
		}
		return errors.New("failed to add tcp frontend")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(addTcpFrontendRes.Body)
	// create bind
	bindParams := QueryParameters{}
	bindParams.add("transaction_id", transactionId)
	bindParams.add("frontend", frontendName)
	bindBody := map[string]interface{}{
		"ssl":  false,
		"port": bindPort,
	}
	bindBodyBytes, err := json.Marshal(bindBody)
	if err != nil {
		return errors.New("failed to marshal bind create request body")
	}
	// Send request
	addBindRes, addBindErr := s.postRequest("/services/haproxy/configuration/binds", bindParams, bytes.NewReader(bindBodyBytes))
	if addBindErr != nil || !isValidStatusCode(addBindRes.StatusCode) {
		return errors.New("failed to add port to frontend")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(addBindRes.Body)
	return nil
}

func (s Manager) IsFrontendExist(transactionId string, listenerMode ListenerMode, bindPort int) (bool, error) {
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	// Send request to check if frontend exist
	isFrontendExistRes, isFrontendExistErr := s.getRequest("/services/haproxy/configuration/frontends/"+frontendName, params)
	if isFrontendExistErr != nil {
		return false, errors.New("failed to check if frontend exist")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(isFrontendExistRes.Body)
	if isFrontendExistRes.StatusCode == 404 {
		return false, nil
	}
	if isFrontendExistRes.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New("failed to check if frontend exist for unknown reason")
}

func (s Manager) IsOtherSwitchingRuleExist(transactionId string, listenerMode ListenerMode, bindPort int) (bool, error) {
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("frontend", frontendName)
	// Send request to fetch switching rules
	switchingRulesRes, switchingRulesErr := s.getRequest("/services/haproxy/configuration/backend_switching_rules", params)
	if switchingRulesErr != nil {
		return false, errors.New("failed to check if switching rule exist")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(switchingRulesRes.Body)
	if switchingRulesRes.StatusCode == 200 {
		switchingRules := map[string]interface{}{}
		if err := json.NewDecoder(switchingRulesRes.Body).Decode(&switchingRules); err != nil {
			return false, errors.New("failed to decode switching rules response")
		}
		switchingRulesArray := switchingRules["data"].([]interface{})
		return len(switchingRulesArray) > 0, nil
	}
	return false, errors.New("failed to check if switching rule exist for unknown reason")
}

func (s Manager) DeleteFrontend(transactionId string, listenerMode ListenerMode, bindPort int) error {
	// we should not delete frontend for port 80 and 443
	if bindPort == 80 || bindPort == 443 {
		return nil
	}
	// check if frontend exists
	isFrontendExist, err := s.IsFrontendExist(transactionId, listenerMode, bindPort)
	if err != nil {
		return err
	}
	if !isFrontendExist {
		return nil
	}
	// ignore for tcp
	if listenerMode == HTTPMode {
		// don't delete frontend if there are switching rules
		isSwitchingRuleExist, err := s.IsOtherSwitchingRuleExist(transactionId, listenerMode, bindPort)
		if err != nil {
			return err
		}
		if isSwitchingRuleExist {
			return nil
		}
	}
	// delete frontend
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	// Send request
	deleteFrontendRes, deleteFrontendErr := s.deleteRequest("/services/haproxy/configuration/frontends/"+frontendName, params)
	if deleteFrontendErr != nil || !isValidStatusCode(deleteFrontendRes.StatusCode) {
		return errors.New("failed to delete frontend")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(deleteFrontendRes.Body)
	return nil
}
