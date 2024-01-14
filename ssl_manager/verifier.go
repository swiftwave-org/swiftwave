package Manager

import (
	"github.com/mrz1836/go-sanitize"
	"io"
	"log"
	"net/http"
	"strings"
)

// VerifyDomain Verify whether the domain is pointing to the server
// Run this before requesting certificate from ACME
func (s Manager) VerifyDomain(domain string) bool {
	finalDomain := "http://" + domain + "/.well-known/pre-authorize/"
	sanitizedDomain, err := sanitize.Domain(finalDomain, false, false)
	if err != nil {
		log.Println("Error sanitizing domain:", err)
		return false
	}
	resp, err := http.Get(sanitizedDomain)
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
