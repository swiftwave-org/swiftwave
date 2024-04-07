package ssh_toolkit

var sshTCPTimeoutSeconds = 10

func UpdateTCPTimeout(t int) {
	sshTCPTimeoutSeconds = t
}
