package Manager

import (
	"io"
	"net/http"
	"strings"
)

// Verify whether the domain is pointing to the server
// Run this before requesting certificate from ACME
func (s Manager) VerifyDomain(domain string) bool {
	// TODO: sanitize domain name
	req, err := http.NewRequest("GET", "http://"+domain+"/.well-known/pre-authorize/", nil)
	if err != nil {
		return false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	// Close response body
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode == 200 {
		// Read response body
		respBody := make([]byte, 1024)
		_, err = resp.Body.Read(respBody)
		if err != nil {
			return false
		} else {
			// Check if the response is "ok"
			return strings.ToLower(string(respBody)) == "ok"
		}
	}
	return false
}
