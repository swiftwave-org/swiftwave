package worker

import (
	"bytes"
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"strings"
	"time"
)

func (m Manager) InstallDependenciesOnServer(request InstallDependenciesOnServerRequest, ctx context.Context, _ context.CancelFunc) error {
	// fetch server
	server, err := core.FetchServerByID(&m.ServiceManager.DbClient, request.ServerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// fetch server log
	serverLog, err := core.FetchServerLogByID(&m.ServiceManager.DbClient, request.LogId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// log
	logText := "Installing dependencies on server\n"
	// spawn a goroutine to update server log each 5 seconds
	go func() {
		lastSent := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if time.Since(lastSent) > 5*time.Second {
					serverLog.Content = logText
					_ = serverLog.Update(&m.ServiceManager.DbClient)
					lastSent = time.Now()
				}
			}
		}
	}()
	// defer to push final log
	defer func() {
		serverLog.Content = logText
		_ = serverLog.Update(&m.ServiceManager.DbClient)
	}()

	detectedOS, err := ssh_toolkit.DetectOS(5, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey, 30)
	if err != nil {
		logText += "Error detecting OS: " + err.Error() + "\n"
		return nil
	}

	// command
	var command string
	for _, dependency := range core.RequiredServerDependencies {
		isExists := false
		// check if dependency is already installed [ignore init]
		if dependency != "init" {
			stdoutBuffer := new(bytes.Buffer)
			err = ssh_toolkit.ExecCommandOverSSH(core.DependencyCheckCommands[dependency], stdoutBuffer, nil, 5, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey, 30)
			if err != nil {
				if strings.Contains(err.Error(), "exited with status 1") {
					isExists = false
				}
			}
			isExists = stdoutBuffer.String() != ""
		}
		// install dependency
		if isExists {
			logText += "Dependency " + dependency + " is already installed\n"
			continue
		} else {
			logText += "Installing dependency " + dependency + "\n"
			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)
			if detectedOS == ssh_toolkit.DebianBased {
				command = core.DebianDependenciesInstallCommands[dependency]
			} else if detectedOS == ssh_toolkit.FedoraBased {
				command = core.FedoraDependenciesInstallCommands[dependency]
			} else {
				logText += "Unknown OS: " + string(detectedOS) + "\n"
				continue
			}
			err = ssh_toolkit.ExecCommandOverSSH(command, stdoutBuffer, stderrBuffer, 5, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey, 30)
			logText += stdoutBuffer.String() + "\n"
			logText += stderrBuffer.String() + "\n"
			logText += "\n"
			if err != nil {
				logText += "Error installing dependency " + dependency + ": " + err.Error() + "\n" + stderrBuffer.String() + "\n"
				return nil
			} else {
				logText += "Dependency " + dependency + " installed successfully\n"
			}
		}
	}
	return nil
}
