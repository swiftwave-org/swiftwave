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
	return ExecCommandOverSSHWithOptions(cmd, stdoutBuf, stderrBuf, sessionTimeoutSeconds, host, port, user, privateKey, true)
}

func ExecCommandOverSSHWithOptions(cmd string,
	stdoutBuf, stderrBuf *bytes.Buffer, sessionTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string, // for ssh client
	validate bool, // if true, will validate if server is online
) error {
	// fetch ssh client
	sshRecord, err := getSSHClientWithOptions(host, port, user, privateKey, validate)
	if err != nil {
		if isErrorWhenSSHClientNeedToBeRecreated(err) {
			DeleteSSHClient(host)
		}
		return err
	}
	// create session
	session, err := getSSHSessionWithTimeout(sshRecord, sessionTimeoutSeconds)
	if err != nil {
		if isErrorWhenSSHClientNeedToBeRecreated(err) {
			DeleteSSHClient(host)
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
	err = runCommandWithTimeout(session, cmd, sshCommandMaxDuration)
	if err != nil {
		if isErrorWhenSSHClientNeedToBeRecreated(err) || isErrorWhenSSHClientNeedToBeRecreated(errors.New(stderrBuf.String())) {
			DeleteSSHClient(host)
			return fmt.Errorf("%s - %s", err.Error(), stderrBuf.String())
		}
		return err
	}
	return nil
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
		// the session may still turn up later, it would leak a channel on the server
		go func() {
			if result := <-resultCh; result.session != nil {
				_ = result.session.Close()
			}
		}()
		return nil, fmt.Errorf("session creation timeout after %d seconds", timeout)
	}
}

// runCommandWithTimeout is a safety net, a command that never finishes should not pin a
// goroutine forever. Detecting a dead connection is the pool keepalive's job.
func runCommandWithTimeout(session *ssh.Session, cmd string, timeout time.Duration) error {
	if err := session.Start(cmd); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		_ = session.Signal(ssh.SIGKILL)
		_ = session.Close()
		return fmt.Errorf("command timeout after %s", timeout)
	}
}
