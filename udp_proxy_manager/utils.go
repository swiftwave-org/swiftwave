package udp_proxy_manager

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
)

// Generate Base URI for HAProxy Server
func (m Manager) URI() string {
	return "http://unix/v1"
}

// Wrapper to send request to HAProxy Server
func (m Manager) getRequest(route string) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = m.URI()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", m.unixSocketPath)
			},
		},
	}
	return client.Do(req)
}

// Wrapper to send request to HAProxy Server
func (m Manager) postRequest(route string, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	var url = m.URI()
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", m.unixSocketPath)
			},
		},
	}
	return client.Do(req)
}
