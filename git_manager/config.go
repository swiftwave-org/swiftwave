package gitmanager

import "sync/atomic"

// insecureSkipTLS controls whether TLS certificate validation is skipped when
// talking to remote Git servers over HTTPS. Defaults to false so credentials
// are never sent over an unverified connection unless explicitly opted in.
var insecureSkipTLS atomic.Bool

// SetInsecureSkipTLS toggles TLS certificate validation for HTTPS remotes.
// Pass true only when intentionally talking to a Git server with a self-signed
// or otherwise non-validatable certificate (e.g., internal test environments).
func SetInsecureSkipTLS(skip bool) {
	insecureSkipTLS.Store(skip)
}

func isInsecureSkipTLS() bool {
	return insecureSkipTLS.Load()
}
