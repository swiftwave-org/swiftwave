package udp_proxy_manager

import (
	"context"
	"io"
	"log"
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
	var url = m.URI() + route
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return m.netConn, nil
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
	var url = m.URI() + route
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return m.netConn, nil
			},
		},
	}
	return client.Do(req)
}

// Close : Close the connection
func (m Manager) Close() {
	err := m.netConn.Close()
	if err != nil {
		log.Println("Error while closing the connection", err)
	}
}

/*
IsPortRestrictedForManualConfig

This function is used to check if a port is restricted or not for application.
There are some ports that are restricted.
because those port are pre-occupied by Swarm services or other required services.
So, binding to those ports will cause errors.
That's why we need to restrict those ports before apply the config.
*/
func IsPortRestrictedForManualConfig(port int, restrictedPorts []int) bool {
	for _, p := range restrictedPorts {
		if port == p {
			return true
		}
	}
	return false
}
