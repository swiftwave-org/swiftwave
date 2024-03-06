#### SSH Toolkit

This is a toolkit for performing various operations in remote servers using SSH.
It will be wrapper around some required host level operations to be performed on remote servers.

**Overall features -**
- Manage Pool of TCP connections to remote servers, so that we can reduce handshake time for each request.
- **ExecCommandOverSSH** - Run any command. Alternative to `exec.Command` in Go.
- **NetConnOverSSH** -  Helper to run `tcp` or `unix` based http requests on remote server. It should return `http.Client` object for further operations, so that `unix` or `tcp` based http requests can be made.
- **CopyFileToRemoteServer** - Copy files to remote server. Use `rsync` for this.
- **CopyFileFromRemoteServer** - Copy files from remote server. Use `rsync` for this.

> [!NOTE]  
> **SSH Toolkit** has a implementation of `Pool of TCP connections` to remote servers. This will help in reducing handshake time for each request.
> The remote servers should be configured to accept `high number of sessions` for single tcp connection.
>
> Go to `/etc/ssh/sshd_config` and set `MaxSessions` to `20` or more. We need to keep it little bit higher than the number of workers configured in swiftwave to keep system stable. Else new session will be timed out and create new TCP connections which is costly and time-consuming.
