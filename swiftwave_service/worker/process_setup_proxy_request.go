package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (m Manager) SetupAndEnableProxy(request SetupAndEnableProxyRequest, ctx context.Context, cancelCtx context.CancelFunc) error {
	// fetch server
	server, err := core.FetchServerByID(&m.ServiceManager.DbClient, request.ServerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.setupAndEnableProxy(request, ctx, cancelCtx)
	if err == nil {
		// mark server as proxy enabled
		server.ProxyConfig.Enabled = true
		server.ProxyConfig.SetupRunning = false
		err = core.UpdateServer(&m.ServiceManager.DbClient, server)
		if err != nil {
			log.Println("Failed to update server:", err)
		} else {

		}
	}
	return nil
}
func (m Manager) setupAndEnableProxy(request SetupAndEnableProxyRequest, ctx context.Context, _ context.CancelFunc) error {
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
	logText := "Starting proxy setup on server " + server.HostName + "\n"
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
	// fill local haproxy configuration (will be skipped anyhow if already exists)
	err = generateDefaultHAProxyConfiguration(m.Config)
	if err != nil {
		logText += "Failed to generate default haproxy configuration: " + err.Error() + "\n"
		return err
	}
	// check if any proxy server is already running
	servers, err := core.FetchAllProxyServers(&m.ServiceManager.DbClient)
	if err != nil {
		logText += "Failed to fetch all proxy servers: " + err.Error() + "\n"
		return err
	}
	if len(servers) > 0 {
		var chosenServer core.Server
		// try to find out an active proxy server
		activeProxyServer, err := core.FetchRandomActiveProxyServer(&m.ServiceManager.DbClient)
		if err == nil {
			chosenServer = activeProxyServer
		} else {
			// if no active proxy server found, choose a random one
			chosenServer = servers[0]
		}
		// copy haproxy directory to the management server
		logText += "Copying haproxy config from server " + chosenServer.HostName + " to local\n"
		err = ssh_toolkit.CopyFolderFromRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, chosenServer.IP, 22, chosenServer.User, m.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			logText += "Failed to copy haproxy config from server " + chosenServer.HostName + " to " + server.HostName + "\n"
			logText += "Error: " + err.Error() + "\n"
			return err
		}
	}
	// copy haproxy directory to the server
	logText += "Copying haproxy config from local to server " + server.HostName + "\n"
	err = ssh_toolkit.CopyFolderToRemoteServer(m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logText += "Failed to copy haproxy config from local to server " + server.HostName + "\n"
		logText += "Error: " + err.Error() + "\n"
		return err
	}

	logText += "Copied haproxy config from local to server " + server.HostName + "\n"
	log.Println("Copied haproxy config from local to server " + server.HostName)
	return nil
}

// generateDefaultHAProxyConfiguration : Generate default haproxy configuration (skip if already exists)
func generateDefaultHAProxyConfiguration(config *config.Config) error {
	// Check if the directory exists
	if _, err := os.Stat(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist > %s", config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath)
	}
	baseUrl, err := generateHAProxyConfigDownloadBaseUrl(config)
	if err != nil {
		return err
	}

	// Check if `haproxy.cfg` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/haproxy.cfg") {
		content, err := downloadContent(baseUrl + "/haproxy.cfg")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `haproxy.cfg` file")
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/haproxy.cfg", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `haproxy.cfg` file")
			}
		}
	}

	// Check if `dataplaneapi.yaml` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/dataplaneapi.yaml") {
		content, err := downloadContent(baseUrl + "/dataplaneapi.yaml")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `dataplaneapi.yaml` file")
			content = strings.ReplaceAll(content, "ADMIN_USERNAME", config.SystemConfig.HAProxyConfig.Username)
			content = strings.ReplaceAll(content, "ADMIN_PASSWORD", config.SystemConfig.HAProxyConfig.Password)
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/dataplaneapi.yaml", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `dataplaneapi.yaml` file")
			}
		}
	}

	// Create `ssl` directory if it does not exist
	if _, err := os.Stat(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/ssl"); os.IsNotExist(err) {
		err := os.MkdirAll(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/ssl", os.ModePerm)
		if err != nil {
			return err
		} else {
			log.Println("Created `ssl` directory")
		}
	}

	// Check if `ssl/default.pem` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/ssl/default.pem") {
		content, err := downloadContent(baseUrl + "/default.pem")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `ssl/default.pem` file")
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/ssl/default.pem", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `ssl/default.pem` file")
			}
		}
	}

	return nil
}

func generateHAProxyConfigDownloadBaseUrl(config *config.Config) (string, error) {
	if config == nil {
		return "", errors.New("config is nil")
	}
	splitString := strings.Split(config.SystemConfig.HAProxyConfig.Image, ":")
	if len(splitString) < 2 {
		return "", errors.New("invalid docker image name")
	}
	version := splitString[1]
	url := "https://raw.githubusercontent.com/swiftwave-org/haproxy/main/" + version
	return url, nil
}

func downloadContent(url string) (string, error) {
	// download with GET request
	res, err := http.Get(url)
	if err != nil {
		return "", errors.New("failed to download file > " + url)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Failed to close response body")
		}
	}(res.Body)

	var fileStringBytes []byte
	// Read the body into bytes
	fileStringBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("failed to read response body")
	}

	// Convert bytes to string
	fileString := string(fileStringBytes)

	return fileString, nil
}

func writeContent(filePath string, content string) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	// Write the content
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	// Close the file
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func checkIfFileExists(file string) bool {
	cmd := exec.Command("ls", file)
	err := cmd.Run()
	return err == nil
}
