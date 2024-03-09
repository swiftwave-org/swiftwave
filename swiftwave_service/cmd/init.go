package cmd

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func init() {
	initCmd.Flags().SortFlags = false
	initCmd.Flags().Bool("auto-domain", false, "Resolve domain name automatically")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SwiftWave configuration on server",
	Run: func(cmd *cobra.Command, args []string) {
		isAutoDomainResolve := cmd.Flag("auto-domain").Value.String() == "true"
		// Try to fetch local config
		_, err := local_config.Fetch()
		if err == nil {
			printError("Config already exists at " + local_config.LocalConfigPath)
			printInfo("Run `swiftwave config` to edit the config file")
			os.Exit(1)
		}

		// Get Domain Name
		domainName, err := resolveQuickDNSDomain()
		if err != nil {
			printError(err.Error())
			if isAutoDomainResolve {
				os.Exit(1)
			}
		}

		if !isAutoDomainResolve {
			// Ask user to enter domain name with default value as resolved domain name
			fmt.Print("Domain Name [default: " + domainName + "]: ")

			var inputDomainName string
			_, err = fmt.Scanln(&inputDomainName)
			if err != nil {
				inputDomainName = ""
			}
			if strings.TrimSpace(inputDomainName) != "" {
				domainName = inputDomainName
				printInfo("Domain name set to " + domainName)
			}
		}
		// Create config
		newConfig := &local_config.Config{
			IsDevelopmentMode: false,
			ServiceConfig: local_config.ServiceConfig{
				UseTLS:                false,
				ManagementNodeAddress: domainName,
			},
			PostgresqlConfig: local_config.PostgresqlConfig{
				Host:                   "127.0.0.1",
				Port:                   5432,
				User:                   generateRandomString(8),
				Password:               generateRandomString(20),
				Database:               "db_" + generateRandomString(8),
				TimeZone:               "Asia/Kolkata",
				SSLMode:                "disable",
				AutoStartLocalPostgres: true,
			},
		}
		err = local_config.FillDefaults(newConfig)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		// generate list of folders to create
		requiredFolders := []string{
			newConfig.ServiceConfig.DataDirectory,
			newConfig.ServiceConfig.LogDirectoryPath,
			newConfig.ServiceConfig.SSLCertDirectoryPath,
			newConfig.ServiceConfig.SocketPathDirectory,
			newConfig.ServiceConfig.HAProxyDataDirectoryPath,
			newConfig.ServiceConfig.UDPProxyDataDirectoryPath,
		}
		// create folders
		for _, folder := range requiredFolders {
			err = createFolder(folder)
			if err != nil {
				printError("Failed to create folder " + folder)
				os.Exit(1)
			}
		}
		// save config
		err = local_config.Update(newConfig)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
	},
}

func resolveQuickDNSDomain() (string, error) {
	// fetch ip address
	ipAddress, err := getIPAddress()
	if err != nil {
		return "", err
	}
	// create domain name
	ipAddress = strings.ReplaceAll(ipAddress, ".", "-")
	quickDNSDomain := fmt.Sprintf("ip-%s.swiftwave.xyz", ipAddress)
	return quickDNSDomain, nil
}

func getIPAddress() (string, error) {
	// send a GET request to https://api.ipify.org/
	resp, err := http.Get("https://api.ipify.org/")
	if err != nil {
		return "", err
	}
	// if response is not 200, return error
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch ip address")
	}
	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp)
	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func generateRandomString(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
