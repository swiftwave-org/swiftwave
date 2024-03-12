package udp_proxy_manager

import (
	"net/http"
)

type Manager struct {
	httpClient *http.Client
}

type Proxy struct {
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Service    string `json:"service"`
}

type AddProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type RemoveProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type ExistProxyResponse struct {
	Exist bool `json:"exist"`
}
