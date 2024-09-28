package cronjob

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
)

func (m Manager) MonitorServerStatus() {
	logger.CronJobLogger.Println("Starting server status monitor [cronjob]")
	for {
		m.monitorServerStatus()
		time.Sleep(10 * time.Second)
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
	}

	var wg sync.WaitGroup
	for _, server := range servers {
		if server.Status == core.ServerNeedsSetup || server.Status == core.ServerPreparing {
			continue
		}
		wg.Add(1)
		go func(server core.Server) {
			defer wg.Done()
			m.checkAndUpdateServerStatus(server)
		}(server)
	}
	wg.Wait()
}

func (m Manager) checkAndUpdateServerStatus(server core.Server) {
	if server.Status == core.ServerOffline {
		ssh_toolkit.DeleteSSHClient(server.HostName)
	}
	if m.isServerOnline(server) {
		err := core.MarkServerAsOnline(&m.ServiceManager.DbClient, &server)
		if err != nil {
			logger.CronJobLoggerError.Println("DB Error : Failed to mark server as online >", server.HostName, err)
		} else {
			logger.CronJobLogger.Println("Server marked as online >", server.HostName)
		}
	} else {
		err := core.MarkServerAsOffline(&m.ServiceManager.DbClient, &server)
		if err != nil {
			logger.CronJobLoggerError.Println("DB Error : Failed to mark server as offline >", server.HostName, err)
		} else {
			logger.CronJobLogger.Println("Server marked as offline >", server.HostName)
		}
	}
}

func (m Manager) isServerOnline(server core.Server) bool {
	// try for 3 times
	for i := 0; i < 3; i++ {
		cmd := "echo ok"
		stdoutBuf := new(bytes.Buffer)
		stderrBuf := new(bytes.Buffer)
		err := ssh_toolkit.ExecCommandOverSSH(cmd, stdoutBuf, stderrBuf, 3, server.IP, server.SSHPort, server.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if strings.Compare(strings.TrimSpace(stdoutBuf.String()), "ok") == 0 {
			return true
		}
	}
	return false
}
