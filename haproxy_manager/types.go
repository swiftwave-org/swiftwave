package haproxymanager

type Manager struct {
	unixSocketPath string
	username       string
	password       string
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
