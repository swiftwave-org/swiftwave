#### SSH Toolkit

This is a toolkit for various operations in remote servers using SSH.
It will be wrapper around some required host level operations to be performed on remote servers.

**Overall features -**
- Manage Pool of TCP connections to remote servers, so that we can reduce handshake time for each request.
- Run any command. Alternative to `exec.Command` in Go.
- Helper to run `tcp` or `unix` based http requests on remote server. It should return `http.Client` object for further operations, so that `unix` or `tcp` based http requests can be made.
- Copy files to remote server. Use `rsync` for this.
- Copy files from remote server. Use `rsync` for this.