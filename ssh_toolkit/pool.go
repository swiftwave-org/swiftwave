package ssh_toolkit

import (
	"errors"
	"fmt"
	"log"
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
	// reject if server is offline
	if validate && sshClientPool.validator != nil && !(*sshClientPool.validator)(host) {
		return nil, errors.New("server is offline, cannot connect to it")
	}
	sshClientPool.mutex.RLock()
	clientEntry, ok := sshClientPool.clients[host]
	sshClientPool.mutex.RUnlock()
	if ok {
		clientEntry.mutex.RLock()
		c := clientEntry.client
		clientEntry.mutex.RUnlock()
		if c != nil {
			return c, nil
		}
	}
	return newSSHClient(host, port, user, privateKey, sshTCPTimeoutSeconds)
}

func newSSHClient(host string, port int, user string, privateKey string, timeoutSeconds int) (*ssh.Client, error) {
	// take pool read lock
	sshClientPool.mutex.RLock()
	// check if another goroutine has created the client in the meantime
	clientEntry, ok := sshClientPool.clients[host]
	// unlock read
	sshClientPool.mutex.RUnlock()
	if ok {
		// take mutex read lock on the entry
		clientEntry.mutex.RLock()
		c := clientEntry.client
		// unlock the mutex on the entry
		clientEntry.mutex.RUnlock()
		if c != nil {
			// another goroutine has created the client
			return c, nil
		}
		// created record but not yet created the client due to handshake in progress + race condition
	}
	// take pool write lock
	sshClientPool.mutex.Lock()
	sshClientRecord := &sshClient{
		client: nil,
		mutex:  &sync.RWMutex{},
	}
	sshClientPool.clients[host] = sshClientRecord
	// take the lock on the new entry,
	// so that no other goroutine can create a client for the same host at the same time
	sshClientRecord.mutex.Lock()
	// release the global lock
	// so that operation for other hosts can continue, as ssh handshake can take time
	sshClientPool.mutex.Unlock()
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		sshClientRecord.mutex.Unlock()
		DeleteSSHClient(host)
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         time.Duration(timeoutSeconds) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		sshClientRecord.mutex.Unlock()
		DeleteSSHClient(host)
		return nil, err
	}
	sshClientRecord.client = client
	sshClientRecord.mutex.Unlock()
	return client, nil
}

func DeleteSSHClient(host string) {
	sshClientPool.mutex.Lock()
	clientEntry, ok := sshClientPool.clients[host]
	if ok {
		clientEntry.mutex.Lock()
		if clientEntry.client != nil {
			err := clientEntry.client.Close()
			if err != nil {
				log.Println("Error closing ssh client:", err)
			}
		}
		clientEntry.mutex.Unlock()
		delete(sshClientPool.clients, host)
	}
	sshClientPool.mutex.Unlock()
}
