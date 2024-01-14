package Manager

import (
	"io"
	"net/http"
	"strings"
)

// VerifyDomain Verify whether the domain is pointing to the server
// Run this before requesting certificate from ACME
func (s Manager) VerifyDomain(domain string) bool {
	finalDomain := "http://" + domain + "/.well-known/pre-authorize/"
	finalDomain = strings.ReplaceAll(finalDomain, "../", "")
	resp, err := http.Get(finalDomain)
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
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		} else {
			// Check if the response is "ok"
			return strings.ToLower(string(respBody)) == "ok"
		}
	}
	return false
}
