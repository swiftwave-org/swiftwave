package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func (m Manager) SetupAndEnableProxy(request SetupAndEnableProxyRequest, ctx context.Context, _ context.CancelFunc) error {
	return nil
}

// try to generate default haproxy configuration
//err := generateDefaultHAProxyConfiguration(m.Config)
//if err != nil {
//logger.CronJobLoggerError.Println("Failed to generate default haproxy configuration", err.Error())
//return
//}

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
