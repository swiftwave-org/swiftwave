package haproxymanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/gommon/random"
	"github.com/tredoe/osutil/user/crypt"
	"github.com/tredoe/osutil/user/crypt/sha256_crypt"
	"io"
	"strconv"
)

func (s Manager) AddUserList(transactionId string, userListName string) error {
	// if already exists, return
	isExists, err := s.IsUserListExist(transactionId, userListName)
	if err != nil {
		return err
	}
	if isExists {
		return nil
	}
	// Build query parameters
	addUserListRequestQueryParams := QueryParameters{}
	addUserListRequestQueryParams.add("transaction_id", transactionId)
	// build request body
	addUserListRequestBody := map[string]interface{}{
		"name": userListName,
	}
	addUserListRequestBodyBytes, err := json.Marshal(addUserListRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_user_list_request_body")
	}
	// Send request to add user list
	userListRes, userListErr := s.postRequest("/services/haproxy/configuration/userlists", addUserListRequestQueryParams, bytes.NewReader(addUserListRequestBodyBytes))
	if userListErr != nil {
		return errors.New("failed to add user list")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to add user list")
	}
	return nil
}

func (s Manager) IsUserListExist(transactionId string, userListName string) (bool, error) {
	// Build query parameters
	isUserListExistRequestQueryParams := QueryParameters{}
	isUserListExistRequestQueryParams.add("transaction_id", transactionId)

	// Send request to check if user list exist
	isUserListExistRes, isUserListExistErr := s.getRequest("/services/haproxy/configuration/userlists/"+userListName, isUserListExistRequestQueryParams)
	if isUserListExistErr != nil {
		return false, errors.New("failed to check if user list exist")
	}
	return isUserListExistRes.StatusCode == 200, nil
}

func (s Manager) DeleteUserList(transactionId string, userListName string) error {
	// Check if user list exist
	isUserListExist, err := s.IsUserListExist(transactionId, userListName)
	if err != nil {
		return err
	}
	if !isUserListExist {
		return nil
	}
	// Build query parameters
	deleteUserListRequestQueryParams := QueryParameters{}
	deleteUserListRequestQueryParams.add("transaction_id", transactionId)
	// Send request to delete user list
	userListRes, userListErr := s.deleteRequest("/services/haproxy/configuration/userlists/"+userListName, deleteUserListRequestQueryParams)
	if userListErr != nil {
		return errors.New("failed to delete user list")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to delete user list")
	}
	return nil
}

func (s Manager) AddUserInUserList(transactionId string, userListName string, username string, encryptedPassword string) error {
	// Check if user exists in user list
	isUserExist, err := s.IsUserExistInUserList(transactionId, userListName, username)
	if err != nil {
		return err
	}
	if isUserExist {
		return errors.New("user already exist in user list")
	}
	// Build query parameters
	addUserInUserListRequestQueryParams := QueryParameters{}
	addUserInUserListRequestQueryParams.add("transaction_id", transactionId)
	addUserInUserListRequestQueryParams.add("userlist", userListName)
	// Add user in user list request body
	addUserInUserListRequestBody := map[string]interface{}{
		"username":        username,
		"secure_password": true,
		"password":        encryptedPassword,
	}
	addUserInUserListRequestBodyBytes, err := json.Marshal(addUserInUserListRequestBody)
	if err != nil {
		return errors.New("failed to marshal add_user_in_user_list_request_body")
	}
	// Send request to add user in user list
	userListRes, userListErr := s.postRequest("/services/haproxy/configuration/users", addUserInUserListRequestQueryParams, bytes.NewReader(addUserInUserListRequestBodyBytes))
	if userListErr != nil {
		return errors.New("failed to add user in user list")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to add user in user list")
	}
	return nil
}

func (s Manager) ChangeUserPasswordInUserList(transactionId string, userListName string, username string, encryptedPassword string) error {
	isUserExist, err := s.IsUserExistInUserList(transactionId, userListName, username)
	if err != nil {
		return err
	}
	if !isUserExist {
		return errors.New("user does not exist in user list")
	}
	// Build query parameters
	changeUserPasswordInUserListRequestQueryParams := QueryParameters{}
	changeUserPasswordInUserListRequestQueryParams.add("transaction_id", transactionId)
	changeUserPasswordInUserListRequestQueryParams.add("userlist", userListName)
	// Change user password in user list request body
	changeUserPasswordInUserListRequestBody := map[string]interface{}{
		"username":        username,
		"secure_password": true,
		"password":        encryptedPassword,
	}
	changeUserPasswordInUserListRequestBodyBytes, err := json.Marshal(changeUserPasswordInUserListRequestBody)
	if err != nil {
		return errors.New("failed to marshal change_user_password_in_user_list_request_body")
	}
	// Send request to change user password in user list
	userListRes, userListErr := s.putRequest("/services/haproxy/configuration/users/"+username, changeUserPasswordInUserListRequestQueryParams, bytes.NewReader(changeUserPasswordInUserListRequestBodyBytes))
	if userListErr != nil {
		return errors.New("failed to change user password in user list")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to change user password in user list")
	}
	return nil
}

func (s Manager) IsUserExistInUserList(transactionId string, userListName string, username string) (bool, error) {
	// Build query parameters
	isUserExistRequestQueryParams := QueryParameters{}
	isUserExistRequestQueryParams.add("transaction_id", transactionId)
	isUserExistRequestQueryParams.add("userlist", userListName)
	// Send request to check if user exist
	isUserExistRes, isUserExistErr := s.getRequest("/services/haproxy/configuration/users/"+username, isUserExistRequestQueryParams)
	if isUserExistErr != nil {
		return false, errors.New("failed to check if user exist")
	}
	return isUserExistRes.StatusCode == 200, nil
}

func (s Manager) DeleteUserFromUserList(transactionId string, userListName string, username string) error {
	isUserExist, err := s.IsUserExistInUserList(transactionId, userListName, username)
	if err != nil {
		return err
	}
	if !isUserExist {
		return nil
	}
	// Build query parameters
	deleteUserInUserListRequestQueryParams := QueryParameters{}
	deleteUserInUserListRequestQueryParams.add("transaction_id", transactionId)
	deleteUserInUserListRequestQueryParams.add("userlist", userListName)
	// Send request to delete user in user list
	userListRes, userListErr := s.deleteRequest("/services/haproxy/configuration/users/"+username, deleteUserInUserListRequestQueryParams)
	if userListErr != nil {
		return errors.New("failed to delete user in user list")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to delete user in user list")
	}
	return nil
}

func createHttpRequestAuthCondition(bindPort int, domain string, userListName string) string {
	rule := fmt.Sprintf("!{ http_auth(%s) } { hdr(host) -i %s }", userListName, domain)
	if bindPort == 80 || bindPort == 443 {
		return rule + " !letsencrypt-acl"
	}
	return rule
}

func (s Manager) SetupBasicAuthentication(transactionId string, listenerMode ListenerMode, bindPort int, domain string, userListName string) error {
	if listenerMode == TCPMode {
		return errors.New("basic authentication is not supported for TCP mode")
	}
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	// check if user-list exists
	isUserListExist, err := s.IsUserListExist(transactionId, userListName)
	if err != nil {
		return err
	}
	if !isUserListExist {
		return errors.New("user list does not exist")
	}
	// add http-request with ACL
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	body := map[string]interface{}{
		"type":      "auth",
		"cond":      "if",
		"cond_test": createHttpRequestAuthCondition(bindPort, domain, userListName),
		"index":     0,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return errors.New("failed to marshal setup_basic_authentication_request_body")
	}
	// Send request to setup basic authentication
	userListRes, userListErr := s.postRequest("/services/haproxy/configuration/http_request_rules", params, bytes.NewReader(bodyBytes))
	if userListErr != nil {
		return errors.New("failed to setup basic authentication")
	}
	if !isValidStatusCode(userListRes.StatusCode) {
		return errors.New("failed to setup basic authentication")
	}
	return nil
}

func (s Manager) RemoveBasicAuthentication(transactionId string, listenerMode ListenerMode, bindPort int, domain string, userListName string) (returnedError error) {
	defer func() {
		// recover from panic
		// added to minimize chance of crash due to unmarshaling of invalid response from haproxy
		if r := recover(); r != nil {
			returnedError = fmt.Errorf("failed to remove basic authentication\npanic: %v", r)
		}
	}()
	if listenerMode == TCPMode {
		return errors.New("basic authentication is not supported for TCP mode")
	}
	frontendName := s.GenerateFrontendName(listenerMode, bindPort)
	// fetch all http-request rules
	params := QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	allHttpRequestRulesRes, allHttpRequestRulesErr := s.getRequest("/services/haproxy/configuration/http_request_rules", params)
	if allHttpRequestRulesErr != nil {
		return errors.New("failed to fetch all http-request rules")
	}
	if !isValidStatusCode(allHttpRequestRulesRes.StatusCode) {
		return errors.New("failed to fetch all http-request rules")
	}
	// read all http-request rules
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(allHttpRequestRulesRes.Body)
	// parse response
	var httpRequestRulesData map[string]interface{}
	err := json.NewDecoder(allHttpRequestRulesRes.Body).Decode(&httpRequestRulesData)
	if err != nil {
		return errors.New("failed to read all http-request rules")
	}
	// find http-request rule
	var httpRequestRule map[string]interface{}
	foundIndex := -1
	for _, rule := range httpRequestRulesData["data"].([]interface{}) {
		httpRequestRule = rule.(map[string]interface{})
		if interfaceToString(httpRequestRule["cond_test"]) == createHttpRequestAuthCondition(bindPort, domain, userListName) &&
			interfaceToString(httpRequestRule["cond"]) == "if" &&
			interfaceToString(httpRequestRule["type"]) == "auth" {
			if httpRequestRule["index"] == nil {
				continue
			}
			foundIndex = int(httpRequestRule["index"].(float64))
			break
		}
	}
	if foundIndex == -1 {
		return nil
	}
	// delete http-request rule
	params = QueryParameters{}
	params.add("transaction_id", transactionId)
	params.add("parent_type", "frontend")
	params.add("parent_name", frontendName)
	deleteHttpRequestRuleRes, deleteHttpRequestRuleErr := s.deleteRequest("/services/haproxy/configuration/http_request_rules/"+strconv.Itoa(foundIndex), params)
	if deleteHttpRequestRuleErr != nil {
		return errors.New("failed to delete http-request rule")
	}
	if !isValidStatusCode(deleteHttpRequestRuleRes.StatusCode) {
		return errors.New("failed to delete http-request rule")
	}
	return nil
}

func GenerateSecuredPasswordForBasicAuthentication(password string) (string, error) {
	c := crypt.New(crypt.SHA256)
	s := sha256_crypt.GetSalt()
	randomSalt := random.String(5)
	saltString := fmt.Sprintf("%s%s", s.MagicPrefix, randomSalt)
	return c.Generate([]byte(password), []byte(saltString))
}

// private function
func interfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}
	return i.(string)
}
