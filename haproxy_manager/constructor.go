package haproxymanager

import (
	"context"
	"net"
	"net/http"
)

// New : Constructor for new instance of haproxy manager
func New(connCreator func() (net.Conn, error), username string, password string) Manager {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return connCreator()
			},
			DisableKeepAlives: true,
		},
	}
	return Manager{
		httpClient: client,
		username:   username,
		password:   password,
	}
}
