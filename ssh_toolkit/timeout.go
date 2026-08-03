package ssh_toolkit

import "time"

var sshTCPTimeoutSeconds = 10

var (
	// keepalive probes, to detect a connection that died without the TCP stack noticing
	sshKeepAliveInterval    = 15 * time.Second
	sshKeepAliveTimeout     = 10 * time.Second
	sshKeepAliveMaxFailures = 3
	// upper bound on a single remote command, so a stuck one can never pin a goroutine forever
	sshCommandMaxDuration = 30 * time.Minute
)

func UpdateTCPTimeout(t int) {
	sshTCPTimeoutSeconds = t
}
