package cronjob

import (
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) SyncBackupProxyServer() {
	for {
		time.Sleep(30 * time.Minute)
		m.syncBackupProxyServer()
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
	err = ssh_toolkit.CopyFileFromRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, activeProxyServer.HostName, 22, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to copy haproxy config from remote server", err.Error())
		return
	}
	// copy udpproxy config to local server
	err = ssh_toolkit.CopyFileFromRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, activeProxyServer.HostName, 22, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to copy udpproxy config from remote server", err.Error())
		return
	}
	// fetch all backup proxy servers
	backupServers, err := core.FetchBackupProxyServers(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch backup proxy servers", err.Error())
		return
	}
	// copy haproxy config to all backup proxy servers
	for _, backupServer := range backupServers {
		err = ssh_toolkit.CopyFileToRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, backupServer.HostName, 22, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy haproxy config to remote server", backupServer.HostName, "\n", err.Error())
		}
	}
	// copy udpproxy config to all backup proxy servers
	for _, backupServer := range backupServers {
		err = ssh_toolkit.CopyFileToRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, backupServer.HostName, 22, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy udpproxy config to remote server", backupServer.HostName, "\n", err.Error())
		}
	}
}
