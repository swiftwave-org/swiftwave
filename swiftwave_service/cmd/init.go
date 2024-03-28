package cmd

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	config2 "github.com/swiftwave-org/swiftwave/swiftwave_service/config"
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
	initCmd.Flags().String("domain", "", "Domain name to use")
	initCmd.Flags().Bool("auto-domain", false, "Resolve domain name automatically")
	initCmd.Flags().Bool("remote-postgres", false, "Opt for remote postgres server")
	initCmd.Flags().Bool("overwrite", false, "Overwrite existing config")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SwiftWave configuration on server",
	Run: func(cmd *cobra.Command, args []string) {
		isAutoDomainResolve := strings.Compare(cmd.Flag("auto-domain").Value.String(), "true") == 0
		isLocalPostgres := strings.Compare(cmd.Flag("remote-postgres").Value.String(), "false") == 0
		isOverWrite := strings.Compare(cmd.Flag("overwrite").Value.String(), "true") == 0
		predefinedDomain := cmd.Flag("domain").Value.String()
		if isOverWrite {
			printWarning("Overwriting existing config! Restart the service to apply changes")
			// try to fetch local config
			val, err := local_config.Fetch()
			if err == nil {
				config = &config2.Config{
					LocalConfig:  val,
					SystemConfig: nil,
				}
			}
		} else {
			// Try to fetch local config
			_, err := local_config.Fetch()
			if err == nil {
				printError("Config already exists at " + local_config.LocalConfigPath)
				printInfo("Run `swiftwave config` to edit the config file")
				os.Exit(1)
			}
		}

		var domainName string

		if strings.Compare(predefinedDomain, "") == 0 {
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
				if strings.Compare(strings.TrimSpace(inputDomainName), "") != 0 {
					domainName = inputDomainName
					printInfo("Domain name set to " + domainName)
				}
			}
		} else {
			printWarning("Using predefined domain name: " + predefinedDomain)
			domainName = predefinedDomain
		}

		currentPostgresHost := ""
		currentPostgresPort := 0
		currentPostgresUser := ""
		currentPostgresPassword := ""
		currentPostgresDatabase := ""
		currentPostgresTimeZone := ""
		currentPostgresSSLMode := ""

		currentLocalImageRegistryImage := ""
		currentLocalImageRegistryPort := 0
		currentLocalImageRegistryUser := ""
		currentLocalImageRegistryPassword := ""

		if config != nil && config.LocalConfig != nil {
			currentPostgresHost = config.LocalConfig.PostgresqlConfig.Host
			currentPostgresPort = config.LocalConfig.PostgresqlConfig.Port
			currentPostgresUser = config.LocalConfig.PostgresqlConfig.User
			currentPostgresPassword = config.LocalConfig.PostgresqlConfig.Password
			currentPostgresDatabase = config.LocalConfig.PostgresqlConfig.Database
			currentPostgresTimeZone = config.LocalConfig.PostgresqlConfig.TimeZone
			currentPostgresSSLMode = config.LocalConfig.PostgresqlConfig.SSLMode
			currentLocalImageRegistryImage = config.LocalConfig.LocalImageRegistryConfig.Image
			currentLocalImageRegistryPort = config.LocalConfig.LocalImageRegistryConfig.Port
			currentLocalImageRegistryUser = config.LocalConfig.LocalImageRegistryConfig.Username
			currentLocalImageRegistryPassword = config.LocalConfig.LocalImageRegistryConfig.Password
		}

		// Create config
		newConfig := &local_config.Config{
			IsDevelopmentMode: false,
			ServiceConfig: local_config.ServiceConfig{
				UseTLS:                      false,
				ManagementNodeAddress:       domainName,
				AutoRenewManagementNodeCert: false,
			},
			PostgresqlConfig: local_config.PostgresqlConfig{
				Host:             defaultString(currentPostgresHost, "127.0.0.1"),
				Port:             defaultInt(currentPostgresPort, 5432),
				User:             defaultString(currentPostgresUser, "user_"+generateRandomString(8)),
				Password:         defaultString(currentPostgresPassword, generateRandomString(20)),
				Database:         defaultString(currentPostgresDatabase, "db_"+generateRandomString(8)),
				TimeZone:         defaultString(currentPostgresTimeZone, "Asia/Kolkata"),
				SSLMode:          defaultString(currentPostgresSSLMode, "disable"),
				RunLocalPostgres: isLocalPostgres,
			},
			LocalImageRegistryConfig: local_config.LocalImageRegistryConfig{
				Image:    defaultString(currentLocalImageRegistryImage, "registry:2.8"),
				Port:     defaultInt(currentLocalImageRegistryPort, 3334),
				Username: defaultString(currentLocalImageRegistryUser, "user_"+generateRandomString(8)),
				Password: defaultString(currentLocalImageRegistryPassword, generateRandomString(20)),
			},
		}
		err := local_config.FillDefaults(newConfig)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		// generate list of folders to create
		requiredFolders := []string{
			newConfig.ServiceConfig.DataDirectory,
			newConfig.ServiceConfig.SocketPathDirectory,
			newConfig.ServiceConfig.TarballDirectoryPath,
			newConfig.ServiceConfig.LogDirectoryPath,
			newConfig.ServiceConfig.PVBackupDirectoryPath,
			newConfig.ServiceConfig.PVRestoreDirectoryPath,
			newConfig.ServiceConfig.SSLCertDirectoryPath,
			newConfig.ServiceConfig.LocalImageRegistryDirectoryPath,
			newConfig.LocalImageRegistryConfig.CertPath,
			newConfig.LocalImageRegistryConfig.AuthPath,
			newConfig.LocalImageRegistryConfig.DataPath,
			newConfig.ServiceConfig.HAProxyDataDirectoryPath,
			newConfig.ServiceConfig.HAProxyUnixSocketDirectory,
			newConfig.ServiceConfig.UDPProxyDataDirectoryPath,
			newConfig.ServiceConfig.UDPProxyUnixSocketDirectory,
			newConfig.ServiceConfig.LocalPostgresDataDirectory,
		}
		// create folders
		for _, folder := range requiredFolders {
			err := createFolder(folder)
			if err != nil {
				printError("Failed to create folder " + folder)
				os.Exit(1)
			} else {
				printSuccess("Folder created > " + folder)
			}
		}
		// save config
		err = local_config.Update(newConfig)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		printSuccess("Config created at " + local_config.LocalConfigPath)
		config = &config2.Config{
			LocalConfig:  newConfig,
			SystemConfig: nil,
		}
		if !isLocalPostgres {
			printInfo("You have opted to use your own postgres server")
			printInfo("Configure postgresql credentials to connect to your own postgres server")
			printInfo("Run `swiftwave config` to edit the config file")
		}
		if isOverWrite {
			printInfo("Config has been overwritten")
			printInfo("Try to restarting the service to apply changes [as overwriting config]")
			restartServiceCmd.Run(serviceCmd, []string{})
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

func defaultString(value, defaultValue string) string {
	if strings.Compare(value, "") == 0 {
		return defaultValue
	}
	return value
}

func defaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}
