package haproxymanager

import (
	"errors"
	"strings"
)

var errInvalidDomainName = errors.New("invalid domain name")

// ValidateDomainName accepts only an RFC 1123 hostname.
//
// Domain names end up inside haproxy ACL expressions and config file names, both of which
// are built by string concatenation, so anything outside [a-zA-Z0-9-.] must be rejected
// before it reaches the shared frontend configuration.
func ValidateDomainName(domainName string) error {
	domainName = strings.TrimSpace(domainName)
	if domainName == "" || len(domainName) > 253 {
		return errInvalidDomainName
	}
	for _, label := range strings.Split(domainName, ".") {
		if !isValidDomainLabel(label) {
			return errInvalidDomainName
		}
	}
	return nil
}

func isValidDomainLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, char := range label {
		isAllowed := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-'
		if !isAllowed {
			return false
		}
	}
	return true
}
