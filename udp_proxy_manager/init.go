package udp_proxy_manager

import (
	"context"
	"net"
	"net/http"
)

// New : Constructor for new instance of udp proxy manager
func New(connCreator func() (net.Conn, error)) Manager {
	return Manager{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return connCreator()
				},
				DisableKeepAlives: true,
			},
		},
	}
}
