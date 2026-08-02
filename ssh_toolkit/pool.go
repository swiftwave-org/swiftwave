package ssh_toolkit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var sshClientPool *sshConnectionPool

func init() {
	sshClientPool = &sshConnectionPool{
		clients:   make(map[string]*sshClient),
		mutex:     &sync.RWMutex{},
		validator: nil,
	}
}

func SetValidator(validator ServerOnlineStatusValidator) {
	sshClientPool.mutex.Lock()
	defer sshClientPool.mutex.Unlock()
	sshClientPool.validator = &validator
}

func getSSHClientWithOptions(host string, port int, user string, privateKey string, validate bool) (*ssh.Client, error) {
	sshClientPool.mutex.RLock()
	validator := sshClientPool.validator
	sshClientPool.mutex.RUnlock()
	// reject if server is offline
	if validate && validator != nil && !(*validator)(host) {
		return nil, errors.New("server is offline, cannot connect to it")
	}
	entry, owned := acquireSSHClientEntry(host, port, user, credentialHash(privateKey))
	if owned {
		entry.connect(host, port, user, privateKey)
	}
	// a dial plus a handshake, each capped at the tcp timeout, with a little slack
	client, err := entry.wait(time.Duration(2*sshTCPTimeoutSeconds+1) * time.Second)
	if err != nil {
		deleteSSHClientEntry(host, entry)
		return nil, err
	}
	return client, nil
}

// acquireSSHClientEntry returns the pooled entry for host, creating a fresh one when there
// is none or when the credentials changed. The bool reports whether the caller owns the
// handshake for the returned entry.
func acquireSSHClientEntry(host string, port int, user string, credHash string) (*sshClient, bool) {
	var stale *sshClient
	sshClientPool.mutex.Lock()
	if entry, ok := sshClientPool.clients[host]; ok {
		if entry.port == port && entry.user == user && entry.credHash == credHash {
			sshClientPool.mutex.Unlock()
			return entry, false
		}
		stale = entry
	}
	entry := &sshClient{
		ready:    make(chan struct{}),
		port:     port,
		user:     user,
		credHash: credHash,
	}
	sshClientPool.clients[host] = entry
	sshClientPool.mutex.Unlock()
	if stale != nil {
		// closing waits for any in-flight handshake, never do that under the pool lock
		go stale.close()
	}
	return entry, true
}

// deleteSSHClientEntry drops entry from the pool, leaving any newer entry for host intact.
func deleteSSHClientEntry(host string, entry *sshClient) {
	sshClientPool.mutex.Lock()
	if current, ok := sshClientPool.clients[host]; ok && current == entry {
		delete(sshClientPool.clients, host)
	}
	sshClientPool.mutex.Unlock()
	go entry.close()
}

func DeleteSSHClient(host string) {
	sshClientPool.mutex.Lock()
	entry, ok := sshClientPool.clients[host]
	if ok {
		delete(sshClientPool.clients, host)
	}
	sshClientPool.mutex.Unlock()
	if ok {
		go entry.close()
	}
}

// private functions
func (s *sshClient) connect(host string, port int, user string, privateKey string) {
	defer close(s.ready)
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		s.err = err
		return
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	timeout := time.Duration(sshTCPTimeoutSeconds) * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		s.err = err
		return
	}
	// ssh.Dial bounds only the tcp dial, the handshake needs a deadline of its own
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		_ = conn.Close()
		s.err = err
		return
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		s.err = err
		return
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = sshConn.Close()
		s.err = err
		return
	}
	s.client = ssh.NewClient(sshConn, chans, reqs)
	s.stopKeepAlive = make(chan struct{})
	go s.keepAlive(host)
}

func (s *sshClient) wait(timeout time.Duration) (*ssh.Client, error) {
	select {
	case <-s.ready:
		if s.err != nil {
			return nil, s.err
		}
		return s.client, nil
	case <-time.After(timeout):
		return nil, errors.New("ssh handshake failed, timed out")
	}
}

func (s *sshClient) close() {
	<-s.ready
	s.closeOnce.Do(func() {
		if s.stopKeepAlive != nil {
			close(s.stopKeepAlive)
		}
		if s.client != nil {
			if err := s.client.Close(); err != nil {
				log.Println("Error closing ssh client:", err)
			}
		}
	})
}

// keepAlive evicts a connection that died silently, a firewall dropping an idle flow or a
// network partition leaves a client that looks healthy but blocks every command forever.
// Closing it also unblocks whatever is already stuck on it.
func (s *sshClient) keepAlive(host string) {
	ticker := time.NewTicker(sshKeepAliveInterval)
	defer ticker.Stop()
	failures := 0
	for {
		select {
		case <-s.stopKeepAlive:
			return
		case <-ticker.C:
			if s.ping() {
				failures = 0
				continue
			}
			failures++
			if failures >= sshKeepAliveMaxFailures {
				log.Printf("ssh keepalive failed %d times, dropping connection to %s", failures, host)
				deleteSSHClientEntry(host, s)
				return
			}
		}
	}
}

func (s *sshClient) ping() bool {
	// SendRequest itself blocks forever on a dead connection
	result := make(chan error, 1)
	go func() {
		_, _, err := s.client.SendRequest("keepalive@openssh.com", true, nil)
		result <- err
	}()
	select {
	case err := <-result:
		return err == nil
	case <-time.After(sshKeepAliveTimeout):
		return false
	}
}

func credentialHash(privateKey string) string {
	sum := sha256.Sum256([]byte(privateKey))
	return hex.EncodeToString(sum[:])
}
