package Manager

import "net/http"

// Verify whether the domain is pointing to the server
// Run this before requesting certificate from ACME
func (s Manager) VerifyDomain(domain string) bool {
	req, err := http.NewRequest("GET", "http://"+domain+"/.well-known/pre-authorize/", nil)
	if err != nil {
		return false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
