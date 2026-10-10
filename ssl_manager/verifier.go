package Manager

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// VerifyDomain Verify whether the domain is pointing to the server
// Run this before requesting certificate from ACME
func (s Manager) VerifyDomain(domain string) bool {
	// Build the verification URL safely. url.Parse + explicit field assignment
	// prevents path traversal sequences (../, ..%2F, ..%2f) injected via the
	// domain argument from changing the request path.
	u := &url.URL{
		Scheme: "http",
		Host:   strings.TrimSpace(domain),
		Path:   "/.well-known/pre-authorize/",
	}
	if u.Host == "" {
		return false
	}
	// Reject hosts that contain control or path characters.
	if strings.ContainsAny(u.Host, "/?#\\") {
		return false
	}
	// Round-trip through Parse to enforce a valid host component.
	parsed, err := url.Parse(u.String())
	if err != nil || parsed.Host != u.Host || parsed.Path != u.Path {
		return false
	}
	// Create a new HTTP client with a timeout
	client := http.Client{
		Timeout: 20 * time.Second,
	}
	// Create a GET request
	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		return false
	}
	// Perform the request with the client
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
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		// Check if the response is "ok"
		return strings.Compare(strings.ToLower(string(respBody)), "ok") == 0
	}
	return false
}
