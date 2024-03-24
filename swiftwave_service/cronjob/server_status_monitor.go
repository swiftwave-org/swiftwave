package cronjob

import (
	"bytes"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"strings"
	"time"
)

func (m Manager) MonitorServerStatus() {
	logger.CronJobLogger.Println("Starting server status monitor [cronjob]")
	for {
		m.monitorServerStatus()
		time.Sleep(5 * time.Minute)
	}
}

func (m Manager) monitorServerStatus() {
	logger.CronJobLogger.Println("Triggering Server Status Monitor Job")
	// Fetch all servers
	servers, err := core.FetchAllServers(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch server list")
		logger.CronJobLoggerError.Println(err)
		return
	}
	if len(servers) == 0 {
		logger.CronJobLogger.Println("Skipping ! No server found")
		return
	} else {
		for _, server := range servers {
			go func(server core.Server) {
				if m.isServerOnline(server) {
					err = core.MarkServerAsOnline(&m.ServiceManager.DbClient, &server)
					if err != nil {
						logger.CronJobLoggerError.Println("DB Error : Failed to mark server as online > ", server.HostName)
					} else {
						logger.CronJobLogger.Println("Server marked as online > ", server.HostName)
					}
				} else {
					err = core.MarkServerAsOffline(&m.ServiceManager.DbClient, &server)
					if err != nil {
						logger.CronJobLoggerError.Println("DB Error : Failed to mark server as offline > ", server.HostName)
					} else {
						logger.CronJobLogger.Println("Server marked as offline > ", server.HostName)
					}
				}
			}(server)
		}
	}
}

func (m Manager) isServerOnline(server core.Server) bool {
	cmd := "echo ok"
	stdoutBuf := new(bytes.Buffer)
	err := ssh_toolkit.ExecCommandOverSSH(cmd, stdoutBuf, nil, 5, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey, 30)
	if err != nil {
		return false
	}
	if strings.Compare(stdoutBuf.String(), "ok") == 0 {
		return false
	}
	return true
}
