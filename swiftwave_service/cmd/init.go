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
)

//go:embed config.standalone.yml
var standaloneConfigSample []byte

//go:embed config.cluster.yml
var clusterConfigSample []byte

func init() {
	initCmd.Flags().SortFlags = false
	initCmd.Flags().Bool("auto-domain", false, "Resolve domain name automatically")
	initCmd.Flags().Bool("overwrite", false, "Overwrite existing configuration")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SwiftWave configuration on server",
	Run: func(cmd *cobra.Command, args []string) {
		isOverwrite := cmd.Flag("overwrite").Value.String() == "true"
		isAutoDomainResolve := cmd.Flag("auto-domain").Value.String() == "true"

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
			_, err = fmt.Scanln(&inputDomainName)
			if err != nil {
				inputDomainName = ""
			}
			if strings.TrimSpace(inputDomainName) != "" {
				domainName = inputDomainName
				printInfo("Domain name set to " + domainName)
			}
		}

		// TODO: create the config file and folders
		// give option, if you like to edit the files
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
