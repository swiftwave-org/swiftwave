package ssh_toolkit

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"io"
	"time"
)

// DirectSSH opens a direct ssh connection to a server
func DirectSSH(
	ctx context.Context, initCol int, initRow int,
	host string, port int, user string, privateKey string) (
	session *ssh.Session, stdin *io.WriteCloser, stdout *io.Reader, stderr *io.Reader, err error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to parse private key")
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         time.Duration(30) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// dial ssh
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to dial ssh")
	}
	// create session
	session, err = client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, nil, nil, nil, errors.New("failed to create session")
	}
	go func(session *ssh.Session) {
		// wait for context cancel
		<-ctx.Done()
		_ = session.Close()
		_ = client.Close()
	}(session)
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", initRow, initCol, modes); err != nil {
		return nil, nil, nil, nil, errors.New("request for pseudo terminal failed")
	}
	// get stdin, stdout, stderr
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stdin")
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stdout")
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stderr")
	}
	// open shell
	err = session.Shell()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to open shell")
	}
	return session, &stdinPipe, &stdoutPipe, &stderrPipe, nil
}

// DirectSSHToContainer opens a direct ssh connection to a container
func DirectSSHToContainer(
	ctx context.Context, initCol int, initRow int, containerId string,
	dockerHost string, host string, port int, user string, privateKey string) (
	session *ssh.Session, stdin *io.WriteCloser, stdout *io.Reader, stderr *io.Reader, err error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to parse private key")
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         time.Duration(30) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// dial ssh
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to dial ssh")
	}
	// create session
	session, err = client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, nil, nil, nil, errors.New("failed to create session")
	}
	go func(session *ssh.Session) {
		// wait for context cancel
		<-ctx.Done()
		_ = session.Close()
		_ = client.Close()
	}(session)
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", initRow, initCol, modes); err != nil {
		return nil, nil, nil, nil, errors.New("request for pseudo terminal failed")
	}
	// get stdin, stdout, stderr
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stdin")
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stdout")
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to get stderr")
	}
	// open shell
	err = session.Start(fmt.Sprintf("DOCKER_HOST=%s docker exec -it %s /bin/sh", dockerHost, containerId))
	if err != nil {
		return nil, nil, nil, nil, errors.New("failed to open shell")
	}
	return session, &stdinPipe, &stdoutPipe, &stderrPipe, nil
}
