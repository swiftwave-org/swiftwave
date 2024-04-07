package ssh_toolkit

import (
	"bytes"
	"errors"
)

func DetectOS(sessionTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string) (OperatingSystem, error) {
	// detect OS
	// debian -  cat /etc/debian_version [any text]
	// fedora - cat /etc/redhat-release [any text]

	stdoutBuf, stderrBuf := new(bytes.Buffer), new(bytes.Buffer)
	err := ExecCommandOverSSH("cat /etc/debian_version", stdoutBuf, stderrBuf, sessionTimeoutSeconds, host, port, user, privateKey)
	if err == nil {
		return DebianBased, nil
	}

	stdoutBuf, stderrBuf = new(bytes.Buffer), new(bytes.Buffer)
	err = ExecCommandOverSSH("cat /etc/redhat-release", stdoutBuf, stderrBuf, sessionTimeoutSeconds, host, port, user, privateKey)
	if err == nil {
		return FedoraBased, nil
	}

	return "", errors.New("unknown OS")
}
