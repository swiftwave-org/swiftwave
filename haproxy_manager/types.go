package haproxymanager

import (
	"net/http"
)

type Manager struct {
	httpClient *http.Client
	username   string
	password   string
}

type QueryParameter struct {
	key   string
	value string
}

type QueryParameters []QueryParameter

type ListenerMode string

const (
	HTTPMode ListenerMode = "http"
	TCPMode  ListenerMode = "tcp"
)

type BackendProtocol string

const (
	HTTPBackend BackendProtocol = "http"
	TCPBackend  BackendProtocol = "tcp"
)
