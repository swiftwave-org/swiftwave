package ssh_toolkit

import (
	"errors"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// hostKeyCallback is the active callback used for verifying SSH server identities.
// Default is InsecureIgnoreHostKey to preserve historical behavior; callers should
// opt into strict verification via SetKnownHostsFile or SetHostKeyCallback.
var (
	hostKeyCallback   ssh.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	hostKeyCallbackMu sync.RWMutex
)

// SetKnownHostsFile configures host key verification using an OpenSSH-format
// known_hosts file. Any subsequent SSH connection will validate the remote
// host key against the entries in the file.
func SetKnownHostsFile(paths ...string) error {
	if len(paths) == 0 {
		return errors.New("at least one known_hosts file path is required")
	}
	cb, err := knownhosts.New(paths...)
	if err != nil {
		return err
	}
	SetHostKeyCallback(cb)
	return nil
}

// SetHostKeyCallback overrides the default host key verification with a custom
// callback. Useful for tests or for plugging in an alternative trust store.
func SetHostKeyCallback(cb ssh.HostKeyCallback) {
	if cb == nil {
		return
	}
	hostKeyCallbackMu.Lock()
	hostKeyCallback = cb
	hostKeyCallbackMu.Unlock()
}

// SetInsecureIgnoreHostKey restores the default (insecure) behavior of skipping
// host key verification. Provided so callers can explicitly opt back in.
func SetInsecureIgnoreHostKey() {
	hostKeyCallbackMu.Lock()
	hostKeyCallback = ssh.InsecureIgnoreHostKey()
	hostKeyCallbackMu.Unlock()
	log.Println("ssh_toolkit: host key verification disabled — connections are vulnerable to MITM")
}

// getHostKeyCallback returns the currently configured HostKeyCallback.
func getHostKeyCallback() ssh.HostKeyCallback {
	hostKeyCallbackMu.RLock()
	defer hostKeyCallbackMu.RUnlock()
	return hostKeyCallback
}
