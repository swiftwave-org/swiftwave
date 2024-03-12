package cronjob

import (
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) SyncBackupProxyServer() {
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
	err = ssh_toolkit.CopyFolderFromRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, activeProxyServer.IP, 22, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to copy haproxy config from remote server", err.Error())
		return
	} else {
		logger.CronJobLogger.Println("Copied haproxy config from remote server", activeProxyServer.HostName)
	}
	// copy udpproxy config to local server
	err = ssh_toolkit.CopyFolderFromRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, activeProxyServer.IP, 22, activeProxyServer.User, m.Config.SystemConfig.SshPrivateKey)
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
		err = ssh_toolkit.CopyFolderToRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, backupServer.IP, 22, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy haproxy config to remote server", backupServer.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Copied haproxy config to remote server", backupServer.HostName)
		}
	}
	// copy udpproxy config to all backup proxy servers
	for _, backupServer := range backupServers {
		err = ssh_toolkit.CopyFolderToRemoteServer(m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath, backupServer.IP, 22, backupServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to copy udpproxy config to remote server", backupServer.HostName, "\n", err.Error())
		} else {
			logger.CronJobLogger.Println("Copied udpproxy config to remote server", backupServer.HostName)
		}
	}

	// TODO: reload backup proxy servers, avoid using swarm_manager
}
