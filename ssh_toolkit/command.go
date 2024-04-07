package ssh_toolkit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func ExecCommandOverSSH(cmd string,
	stdoutBuf, stderrBuf *bytes.Buffer, sessionTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string, // for ssh client
) error {
	// fetch ssh client
	sshRecord, err := getSSHClient(host, port, user, privateKey)
	if err != nil {
		return err
	}
	// create session
	session, err := getSSHSessionWithTimeout(sshRecord, sessionTimeoutSeconds)
	if err != nil {
		if isErrorWhenSSHClientNeedToBeRecreated(err) {
			deleteSSHClient(host)
		}
		return err
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("Error closing session:", err)
		}
	}(session)
	// set buffers
	if stdoutBuf == nil {
		stdoutBuf = new(bytes.Buffer)
	}
	if stderrBuf == nil {
		stderrBuf = new(bytes.Buffer)
	}
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
