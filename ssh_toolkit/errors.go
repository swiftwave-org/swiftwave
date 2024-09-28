package ssh_toolkit

import (
	"strings"
)

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
	"eof",
	"broken pipe",
	"closing write end of pipe",
	"connection reset by peer",
	"unexpected packet in response to channel open",
}

func isErrorWhenSSHClientNeedToBeRecreated(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	for _, msg := range errorsWhenSSHClientNeedToBeRecreated {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}
	return false
}
