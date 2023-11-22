package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/system_config"
	"gopkg.in/yaml.v3"
)

//go:embed config.standalone.yml
var standaloneConfigSample []byte

//go:embed config.cluster.yml
var clusterConfigSample []byte

func init() {
	initCmd.Flags().SortFlags = false
	initCmd.Flags().String("mode", "standalone", "Mode of operation [standalone or cluster]")
	initCmd.Flags().String("letsencrypt-email", "", "Email address for Let's Encrypt")
	initCmd.Flags().String("haproxy-user", "admin", "Username for HAProxy stats page")
	initCmd.Flags().String("haproxy-password", "admin", "Password for HAProxy stats page")
	initCmd.Flags().Bool("auto-domain", false, "Resolve domain name automatically")
	initCmd.Flags().Bool("overwrite", false, "Overwrite existing configuration")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SwiftWave configuration on server",
	Run: func(cmd *cobra.Command, args []string) {
		isOverwrite := cmd.Flag("overwrite").Value.String() == "true"
		isAutoDomainResolve := cmd.Flag("auto-domain").Value.String() == "true"
		letsEncryptEmail := cmd.Flag("letsencrypt-email").Value.String()
		haproxyUser := cmd.Flag("haproxy-user").Value.String()
		haproxyPassword := cmd.Flag("haproxy-password").Value.String()

		// ensure if git is installed
		if !checkIfCommandExists("git") {
			// exit with message
			printError("Git is not installed on your system. Please install git and try again.")
			return
		}

		// ensure if docker is installed
		if !checkIfCommandExists("docker") {
			// exit with message
			printError("Docker is not installed on your system. Please install docker and try again.")
			printInfo("You can use the following command to install docker on your linux system : ")
			printInfo("curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh")
			printInfo("For more information, visit https://docs.docker.com/engine/install/")
			return
		}

		// Create folder at /etc/swiftwave if not exists
		if !checkIfFolderExists("/etc/swiftwave") {
			err := createFolder("/etc/swiftwave")
			if err != nil {
				printError(err.Error())
				os.Exit(1)
			} else {
				printSuccess("Folder /etc/swiftwave created")
			}
		} else {
			printSuccess("Folder /etc/swiftwave already exists")
		}

		// Create folder at /var/lib/swiftwave if not exists
		if !checkIfFolderExists("/var/lib/swiftwave") {
			err := createFolder("/var/lib/swiftwave")
			if err != nil {
				printError(err.Error())
				os.Exit(1)
			} else {
				printSuccess("Folder /var/lib/swiftwave created")
			}
		} else {
			printSuccess("Folder /var/lib/swiftwave already exists")
		}

		// Create folder at /var/lib/swiftwave/haproxy if not exists
		if !checkIfFolderExists("/var/lib/swiftwave/haproxy") {
			err := createFolder("/var/lib/swiftwave/haproxy")
			if err != nil {
				printError(err.Error())
				os.Exit(1)
			} else {
				printSuccess("Folder /var/lib/swiftwave/haproxy created")
			}
		} else {
			printSuccess("Folder /var/lib/swiftwave/haproxy already exists")
		}

		// Check if config file exists > /etc/swiftwave/config.yml
		if checkIfFileExists(configFilePath) {
			printSuccess("Config file /etc/swiftwave/config.yml already exists")
			// If exists, check if overwrite flag is set
			if isOverwrite {
				// If yes, prompt user to overwrite
				printSuccess("The operation will overwrite existing config file")
				fmt.Print("Do you want to continue? [y/N] ")
				var response string
				_, err := fmt.Scanln(&response)
				if err != nil {
					log.Println(err.Error())
					os.Exit(1)
				}
				if !(response == "y" || response == "Y") {
					return
				}
			} else {
				printError("Config file already exists. Use --overwrite flag to overwrite existing config file")
				os.Exit(1)
			}
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
			_, err = fmt.Scanln(&inputDomainName, "")
			if err != nil {
				inputDomainName = ""
			}
			if strings.TrimSpace(inputDomainName) != "" {
				domainName = inputDomainName
				printInfo("Domain name set to " + domainName)
			}
		}

		// Ask user to enter email address for Let's Encrypt
		if strings.TrimSpace(letsEncryptEmail) == "" {
			fmt.Print("Enter Email Address (will be used for LetsEncrypt): ")
			_, err = fmt.Scanln(&letsEncryptEmail)
			if err != nil {
				letsEncryptEmail = ""
			}
			if strings.TrimSpace(letsEncryptEmail) == "" {
				printError("Email address is required for Let's Encrypt. Retry !")
				os.Exit(1)
			}
		}

		// If not exists, create config file
		mode := cmd.Flag("mode").Value.String()
		// create config pointer object
		var configTemplate system_config.Config
		var isCreated bool = false
		if mode == string(system_config.Standalone) {
			err = yaml.Unmarshal(standaloneConfigSample, &configTemplate)
			if err != nil {
				printError("failed to unmarshal standalone config template, please report the issue in github")
				os.Exit(1)
			}
			isCreated = true
		} else if mode == string(system_config.Cluster) {
			err = yaml.Unmarshal(clusterConfigSample, &configTemplate)
			if err != nil {
				printError("failed to unmarshal cluster config template, please report the issue in github")
				os.Exit(1)
			}
		} else {
			printError("Invalid mode of operation. Use --mode flag to specify mode of operation")
			return
		}

		configTemplate.ServiceConfig.AddressOfCurrentNode = domainName
		configTemplate.LetsEncryptConfig.EmailID = letsEncryptEmail
		configTemplate.HAProxyConfig.User = haproxyUser
		configTemplate.HAProxyConfig.Password = haproxyPassword

		isCreated = createConfig(configTemplate, configFilePath)

		if isCreated {
			printSuccess("Config file created at /etc/swiftwave/config.yml")
			printInfo("Run `swiftwave config` to open the config file in editor")
			var msg = ""
			if mode == "standalone" {
				msg =
					`You need to edit at-least the following parameters in the config file:
- Postgres database credentials (If you want to use external database)
`
			} else if mode == "cluster" {
				msg =
					`You need to edit at-least the following parameters in the config file:
- Postgres database credentials
- Redis database credentials
- RabbitMQ credentials`
			}
			fmt.Println(msg)
		} else {
			printError("Failed to create config file at /etc/swiftwave/config.yml")
			os.Exit(1)
		}
	},
}

func createConfig(config system_config.Config, configFilePath string) bool {
	// open and write to file
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return false
	}
	// Truncate the file
	err = file.Truncate(0)
	if err != nil {
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			printError(err.Error())
		}
	}(file)
	yamlMarshalledBytes, err := yaml.Marshal(config)
	if err != nil {
		return false
	}
	_, err = file.Write(yamlMarshalledBytes)
	return err == nil
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
