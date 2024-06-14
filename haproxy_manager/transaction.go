package haproxymanager

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
)

// fetchVersion : Fetch HAProxy current configuration version
func (s Manager) fetchVersion() (string, error) {
	res, err := s.getRequest("/services/haproxy/configuration/version", QueryParameters{})
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(res.Body)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	version := strings.TrimSpace(string(bodyBytes))
	return version, nil
}

// FetchNewTransactionId : Generate new transaction id
func (s Manager) FetchNewTransactionId() (string, error) {
	version, err := s.fetchVersion()
	if err != nil {
		return "", errors.New("Error while fetching HAProxy version: " + err.Error())
	}
	queryParams := QueryParameters{}
	queryParams.add("version", version)
	res, err := s.postRequest("/services/haproxy/transactions", queryParams, nil)
	if err != nil || !isValidStatusCode(res.StatusCode) {
		return "", errors.New("failed to fetch version")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] FetchNewTransactionId: ", err)
		}
	}(res.Body)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("invalid response body")
	}
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return "", errors.New("failed to decode json")
	}
	transactionId := data["id"].(string)
	return transactionId, nil
}

// CommitTransaction : Commit new transaction with force reload to apply changes
func (s Manager) CommitTransaction(transactionId string) error {
	queryParams := QueryParameters{}
	queryParams.add("force_reload", "true")
	res, err := s.putRequest("/services/haproxy/transactions/"+transactionId, queryParams, nil)
	if err != nil || !isValidStatusCode(res.StatusCode) {
		return errors.New("error while committing transaction")
	}
	return nil
}

// DeleteTransaction : Delete transaction
func (s Manager) DeleteTransaction(transactionId string) error {
	res, err := s.deleteRequest("/services/haproxy/transactions/"+transactionId, QueryParameters{})
	if err != nil || !isValidStatusCode(res.StatusCode) {
		return errors.New("error while deleting transaction: " + transactionId)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("[haproxy_manager] DeleteTransaction: ", err)
		}
	}(res.Body)
	return nil
}
