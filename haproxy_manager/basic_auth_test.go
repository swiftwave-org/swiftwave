package haproxymanager

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBasicAuthentication(t *testing.T) {
	t.Run("create user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, "test")
		assert.NoError(t, err, "add user list should not return error")

		config := fetchConfig(transactionId)
		assert.Contains(t, config, "userlist test", "`userlist test` should be present in config")
	})

	t.Run("create duplicate user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, "test")
		assert.NoError(t, err, "add user list should not return error")
		assert.Contains(t, fetchConfig(transactionId), "userlist test", "`userlist test` should be present in config")

		err = haproxyTestManager.AddUserList(transactionId, "test")
		assert.NoError(t, err, "add user list should not return error")
		assert.Equal(t, strings.Count(fetchConfig(transactionId), "userlist test"), 1, "`userlist test` should be present only once in config")
	})

	t.Run("is exits user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		userListName := "test"

		isExist, err := haproxyTestManager.IsUserListExist(transactionId, userListName)
		assert.NoError(t, err, "check if user list exist should not return error")
		assert.False(t, isExist, "check if user list exist should be false")

		err = haproxyTestManager.AddUserList(transactionId, "test")
		assert.NoError(t, err, "add user list should not return error")

		exists, err := haproxyTestManager.IsUserListExist(transactionId, "test")
		assert.NoError(t, err, "check if user list exist should not return error")
		assert.True(t, exists, "check if user list exist should be true")
	})

	t.Run("delete user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, "test")
		assert.NoError(t, err, "add user list should not return error")

		config := fetchConfig(transactionId)
		assert.Contains(t, config, "userlist test", "`userlist test` should be present in config")

		err = haproxyTestManager.DeleteUserList(transactionId, "test")
		assert.NoError(t, err, "delete user list should not return error")

		exists, err := haproxyTestManager.IsUserListExist(transactionId, "test")
		assert.NoError(t, err, "check if user list exist should not return error")
		assert.False(t, exists, "check if user list exist should be false")
	})

	t.Run("delete non_exist user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.DeleteBackend(transactionId, "test")
		assert.NoError(t, err, "delete non-exist user list should not return error")
	})

	userListName := "test"
	username := "usera"
	password := "passworda"
	protectedDomain := "example.com"

	t.Run("add user to user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		err = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		assert.NoError(t, err, "add user list should not return error")

		config := fetchConfig(transactionId)
		searchString := fmt.Sprintf("user %s password", username)
		assert.Contains(t, config, searchString, "user should be present in config")
	})

	t.Run("adding duplicate user to user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		err = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		assert.NoError(t, err, "add user in list should not return error")

		err = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		assert.Error(t, err, "add duplicate user in list should return error")
	})

	t.Run("check if user exists in user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		isExist, err := haproxyTestManager.IsUserExistInUserList(transactionId, userListName, username)
		assert.NoError(t, err, "check if user exist in user list should not return error")
		assert.False(t, isExist, "check if user exist in user list should be false")

		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		err = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		assert.NoError(t, err, "add user in user list should not return error")

		isExist, err = haproxyTestManager.IsUserExistInUserList(transactionId, userListName, username)
		assert.NoError(t, err, "check if user exist in user list should not return error")
		assert.True(t, isExist, "check if user exist in user list should be true")
	})

	t.Run("delete user from user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		err = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		assert.NoError(t, err, "add user in user list should not return error")

		isExist, err := haproxyTestManager.IsUserExistInUserList(transactionId, userListName, username)
		assert.NoError(t, err, "check if user exist in user list should not return error")
		assert.True(t, isExist, "check if user exist in user list should be true")

		err = haproxyTestManager.DeleteUserFromUserList(transactionId, userListName, username)
		assert.NoError(t, err, "delete user in user list should not return error")

		isExist, err = haproxyTestManager.IsUserExistInUserList(transactionId, userListName, username)
		assert.NoError(t, err, "check if user exist in user list should not return error")
		assert.False(t, isExist, "check if user exist in user list should be false")
	})

	t.Run("delete non-exist user from user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		err = haproxyTestManager.DeleteUserFromUserList(transactionId, userListName, username)
		assert.NoError(t, err, "delete non-exist user in user list should not return error")
	})

	t.Run("delete non-exist user from non-exist user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.DeleteUserFromUserList(transactionId, userListName, username)
		assert.NoError(t, err, "delete non-exist user from non-exist user list should not return error")
	})

	t.Run("update user password in user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)

		config := fetchConfig(transactionId)
		searchString := fmt.Sprintf("user %s password", username)
		assert.Contains(t, config, searchString, "user should be present in config")

		newPassword := "newpassword"
		encryptedNewPassword, _ := GenerateSecuredPasswordForBasicAuthentication(newPassword)
		err := haproxyTestManager.ChangeUserPasswordInUserList(transactionId, userListName, username, encryptedNewPassword)
		assert.NoError(t, err, "change user password in user list should not return error")

		config = fetchConfig(transactionId)
		searchString = fmt.Sprintf("user %s password", username)
		assert.Contains(t, config, searchString, "user should be present in config")
	})

	t.Run("update non-exist user in user list", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		err := haproxyTestManager.AddUserList(transactionId, userListName)
		assert.NoError(t, err, "add user list should not return error")

		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		err = haproxyTestManager.ChangeUserPasswordInUserList(transactionId, userListName, username, encryptedPassword)
		assert.Error(t, err, "updating non-exist user's password in user list should return error")
	})

	t.Run("setup basic authentication on port 80", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)

		err := haproxyTestManager.SetupBasicAuthentication(transactionId, HTTPMode, 80, protectedDomain, userListName)
		assert.NoError(t, err, "setup basic authentication should not return error")

		config := fetchConfig(transactionId)
		conditionString := fmt.Sprintf("http-request auth if !{ http_auth(%s) } { hdr(host) -i %s } !letsencrypt-acl", userListName, protectedDomain)
		assert.Contains(t, config, conditionString, "http-request auth condition with !letsencrypt-acl should be present in config")
	})

	t.Run("setup basic authentication on port 443", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)

		err := haproxyTestManager.SetupBasicAuthentication(transactionId, HTTPMode, 443, protectedDomain, userListName)
		assert.NoError(t, err, "setup basic authentication should not return error")

		config := fetchConfig(transactionId)
		conditionString := fmt.Sprintf("http-request auth if !{ http_auth(%s) } { hdr(host) -i %s } !letsencrypt-acl", userListName, protectedDomain)
		assert.Contains(t, config, conditionString, "http-request auth condition with !letsencrypt-acl should be present in config")
	})

	t.Run("setup basic authentication on any port other than 80 and 443", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		_ = haproxyTestManager.AddFrontend(transactionId, HTTPMode, 8080, []int{})

		err := haproxyTestManager.SetupBasicAuthentication(transactionId, HTTPMode, 8080, protectedDomain, userListName)
		assert.NoError(t, err, "setup basic authentication should not return error")

		config := fetchConfig(transactionId)
		conditionString := fmt.Sprintf("http-request auth if !{ http_auth(%s) } { hdr(host) -i %s } !letsencrypt-acl", userListName, protectedDomain)
		assert.NotContains(t, config, conditionString, "http-request auth condition with !letsencrypt-acl should not be present in config for any port other than 80 and 443")

		conditionString = fmt.Sprintf("http-request auth if !{ http_auth(%s) } { hdr(host) -i %s }", userListName, protectedDomain)
		assert.Contains(t, config, conditionString, "http-request auth condition without !letsencrypt-acl should be present in config")
	})

	t.Run("adding basic authentication in tcp mode should raise error", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)

		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)

		err := haproxyTestManager.SetupBasicAuthentication(transactionId, TCPMode, 8080, protectedDomain, userListName)
		assert.Error(t, err, "basic authentication is not supported for TCP mode")
	})

	t.Run("remove basic authentication", func(t *testing.T) {
		transactionId := newTransaction()
		defer deleteTransaction(transactionId)
		// add user list
		_ = haproxyTestManager.AddUserList(transactionId, userListName)
		encryptedPassword, _ := GenerateSecuredPasswordForBasicAuthentication(password)
		_ = haproxyTestManager.AddUserInUserList(transactionId, userListName, username, encryptedPassword)
		_ = haproxyTestManager.SetupBasicAuthentication(transactionId, HTTPMode, 80, protectedDomain, userListName)
		// check if user list exist
		conditionString := fmt.Sprintf("http-request auth if !{ http_auth(%s) } { hdr(host) -i %s } !letsencrypt-acl", userListName, protectedDomain)
		config := fetchConfig(transactionId)
		assert.Contains(t, config, conditionString, "http-request auth condition with !letsencrypt-acl should be present in config")
		// remove basic authentication
		err := haproxyTestManager.RemoveBasicAuthentication(transactionId, HTTPMode, 80, protectedDomain, userListName)
		assert.NoError(t, err, "remove basic authentication should not return error")
		// check if user list exist
		config = fetchConfig(transactionId)
		assert.NotContains(t, config, conditionString, "http-request auth condition with !letsencrypt-acl should not be present in config")
	})
}
