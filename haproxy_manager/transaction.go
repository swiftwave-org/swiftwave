package haproxymanager

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// Fetch HAProxy current configuration version
func (s HAProxySocket) fetchVersion() (string, error){
	res, err := s.getRequest("/services/haproxy/configuration/version", QueryParameters{})
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	version := strings.TrimSpace(string(bodyBytes))
	return version, nil
}

// Generate new transaction id
func (s HAProxySocket) fetchNewTransactionId()(string, error){
	version, err := s.fetchVersion()
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	queryParams := QueryParameters{}
	queryParams.add("version", version)
	res, err := s.postRequest("/services/haproxy/transactions", queryParams, nil)
	if err != nil {
		return "", errors.New("failed to fetch version")
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("invalid response body")
	}
	var data map[string]interface{};
	err = json.Unmarshal(bodyBytes, &data);
	if err != nil {
		return "", errors.New("failed to decode json")
	}
	transactionId := data["id"].(string)
	return transactionId, nil
}

// Commit new transaction with force reload to apply changes
func (s HAProxySocket) commitTransaction(transactionId string) error{
	queryParams := QueryParameters{}
	queryParams.add("force_reload", "true")
	res, err := s.putRequest("/services/haproxy/transactions/"+transactionId, queryParams, nil)
	if err != nil {
		return errors.New("error while committing transaction: "+transactionId)
	}
	defer res.Body.Close()
	return nil
}