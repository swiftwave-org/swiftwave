package ssh_toolkit

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"strings"
	"time"
)

func ExecCommandOverSSH(cmd string,
	stdoutBuf, stderrBuf *bytes.Buffer, sessionTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string, tcpTimeoutSeconds int, // for ssh client
) error {
	// fetch ssh client
	sshRecord, err := getSSHClient(host, port, user, privateKey, tcpTimeoutSeconds)
	if err != nil {
		return err
	}
	// create session
	session, err := getSSHSessionWithTimeout(sshRecord, sessionTimeoutSeconds)
	if err != nil {
		if strings.Contains(err.Error(), "session creation timeout") {
			deleteSSHClient(host)
		}
		return err
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {
			log.Println("Error closing session:", err)
		}
	}(session)
	// set buffers
	session.Stdout = stdoutBuf
	session.Stderr = stderrBuf
	// run command
	return session.Run(cmd)
}

// private functions
func getSSHSessionWithTimeout(client *ssh.Client, timeout int) (*ssh.Session, error) {
	type sessionResult struct {
		session *ssh.Session
		err     error
	}
	resultCh := make(chan sessionResult, 1)
	go func() {
		session, err := client.NewSession()
		resultCh <- sessionResult{session, err}
	}()
	select {
	case result := <-resultCh:
		return result.session, result.err
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, fmt.Errorf("session creation timeout after %d seconds", timeout)
	}
}
