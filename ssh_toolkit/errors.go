package ssh_toolkit

import "strings"

var errorsWhenSSHClientNeedToBeRecreated = []string{
	"dial timeout",
	"i/o timeout",
	"session creation timeout",
	"failed to dial ssh",
	"handshake failed",
	"unable to authenticate",
	"rejected: too many authentication failures",
	"rejected: connection closed by remote host",
	"rejected: connect failed",
	"open failed",
	"handshake failed",
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
