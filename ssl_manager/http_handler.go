package Manager

import (
	"io"
	"net/http"
	"strings"
)

// Required for http-01 verification
// - Path /.well-known/acme-challenge/{token}
func (s Manager) ACMEHttpHandler(w http.ResponseWriter, r *http.Request) error {
	token := strings.ReplaceAll(r.URL.Path, "/.well-known/acme-challenge/", "")
	fullToken := s.fetchKeyAuthorization(token)
	_, err := io.WriteString(w, fullToken)
	return err
}

// Required for pre-authorization
// Check if the domain is pointing to the server
// - Path /.well-known/pre-authorize/
func (s Manager) DNSConfigurationPreAuthorizeHttpHandler(w http.ResponseWriter, r *http.Request) error {
	_, err := io.WriteString(w, "OK")
	return err
}
