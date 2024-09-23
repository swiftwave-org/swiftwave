package ssh_toolkit

import "strings"

var errorsWhenSSHClientNeedToBeRecreated = []string{
	"dial timeout",
	"i/o timeout",
	"session creation timeout",
	"session failed",
	"failed to dial",
	"unable to dial",
	"handshake failed",
	"password auth failed",
	"keyboard-interactive failed",
	"unable to authenticate",
	"server got error",
	"client could not authenticate",
	"connection refused",
	"use of closed network connection",
	"many authentication failures",
	"connection closed by remote host",
	"connect failed",
	"open failed",
	"handshake failed",
	"subsystem request failed",
	"EOF",
	"broken pipe",
	"closing write end of pipe"
}

func isErrorWhenSSHClientNeedToBeRecreated(err error) bool {
	if err == nil {
		return false
	}
	for _, msg := range errorsWhenSSHClientNeedToBeRecreated {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}
