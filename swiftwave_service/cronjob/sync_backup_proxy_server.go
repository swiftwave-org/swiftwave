package cronjob

import (
	"context"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) SyncBackupProxyServer() {
	logger.CronJobLogger.Println("Starting sync backup proxy server [cronjob]")
	for {
		m.syncBackupProxyServer()
		time.Sleep(30 * time.Minute)
	}
}

func (m Manager) syncBackupProxyServer() {
	// Pick any active proxy server
	activeProxyServer, err := core.FetchRandomActiveProxyServer(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch random active proxy server", err.Error())
		return
	}
	// copy haproxy config to local server
	err = ssh_toolkit.CopyFolderFromRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, activeProxyServer.IP, activeProxyServer.SSHPort, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to copy haproxy config from remote server", err.Error())
		return
	} else {
		logger.CronJobLogger.Println("Copied haproxy config from remote server", activeProxyServer.HostName)
	}
	// copy udpproxy config to local server
	err = ssh_toolkit.CopyFolderFromRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, activeProxyServer.IP, activeProxyServer.SSHPort, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to copy udpproxy config from remote server", err.Error())
		return
	} else {
		logger.CronJobLogger.Println("Copied udpproxy config from remote server", activeProxyServer.HostName)
	}
	// fetch all backup proxy servers
	backupServers, err := core.FetchBackupProxyServers(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch backup proxy servers", err.Error())
		return
	}
	// copy haproxy config to all backup proxy servers
	for _, backupServer := range backupServers {
		err = ssh_toolkit.CopyFolderToRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, backupServer.IP, backupServer.SSHPort, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy haproxy config to remote server", backupServer.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Copied haproxy config to remote server", backupServer.HostName)
		}
	}
	// copy udpproxy config to all backup proxy servers
	for _, backupServer := range backupServers {
		err = ssh_toolkit.CopyFolderToRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, backupServer.IP, backupServer.SSHPort, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy udpproxy config to remote server", backupServer.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Copied udpproxy config to remote server", backupServer.HostName)
		}
	}

	// reload proxies on backup server
	for _, server := range backupServers {
		// open ssh connection to backup proxy server for docker
		conn, err := ssh_toolkit.NetConnOverSSH("unix", server.DockerUnixSocketPath, 5, server.IP, server.SSHPort, server.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to open ssh connection to backup proxy server", server.HostName, "\n", err.Error())
			continue
		}
		dockerManager, err := containermanger.New(context.Background(), conn)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to create docker manager for backup proxy server", server.HostName, "\n", err.Error())
			continue
		}
		// remove udpproxy containers from all backup proxy servers, to trigger reload
		err = dockerManager.RemoveServiceContainers(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to remove udpproxy containers from backup proxy server", server.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Removed udpproxy containers from backup proxy server", server.HostName, " for a force reload")
		}
		// reload haproxy on backup proxy serverskill -SIGUSR2 1
		err = dockerManager.RunCommandInServiceContainers(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName, []string{"kill", "-SIGUSR2", "1"})
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to reload haproxy on backup proxy server", server.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Reloaded haproxy on backup proxy server", server.HostName)
		}
		err = conn.Close()
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to close ssh connection to backup proxy server", server.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Closed ssh connection to backup proxy server", server.HostName)
		}
	}
}
