package ssh_toolkit

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var sshClientPool *sshConnectionPool

func init() {
	sshClientPool = &sshConnectionPool{
		clients: make(map[string]*sshClient),
		mutex:   &sync.RWMutex{},
	}
}

func getSSHClient(host string, port int, user string, privateKey string, timeoutSeconds int) (*ssh.Client, error) {
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
	return newSSHClient(host, port, user, privateKey, timeoutSeconds)
}

func newSSHClient(host string, port int, user string, privateKey string, timeoutSeconds int) (*ssh.Client, error) {
	// create entry first with a write lock
	sshClientPool.mutex.Lock()
	// check if another goroutine has created the client in the meantime
	clientEntry, ok := sshClientPool.clients[host]
	if ok {
		sshClientPool.mutex.Unlock()
		clientEntry.mutex.RLock()
		c := clientEntry.client
		clientEntry.mutex.RUnlock()
		return c, nil
	}
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
		deleteSSHClient(host)
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
		deleteSSHClient(host)
		return nil, err
	}
	sshClientRecord.client = client
	sshClientRecord.mutex.Unlock()
	return client, nil
}

func deleteSSHClient(host string) {
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
	}
	delete(sshClientPool.clients, host)
	sshClientPool.mutex.Unlock()
}
