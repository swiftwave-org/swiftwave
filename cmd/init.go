package cmd

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

//go:embed config.standalone.yml
var standaloneConfigSample []byte

//go:embed config.cluster.yml
var clusterConfigSample []byte

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("mode", "standalone", "Mode of operation [standalone or cluster]")
	initCmd.Flags().Bool("overwrite", false, "Overwrite existing configuration [true or false]")
	initCmd.Flags().StringP("editor", "e", "", "Editor to use")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SwiftWave configuration on server",
	Run: func(cmd *cobra.Command, args []string) {
		isOverwrite := cmd.Flag("overwrite").Value.String() == "true"
		// ensure if vim is installed
		if !checkIfCommandExists("vim") {
			// exit with message
			printError("Vim is not installed on your system. Please install vim and try again.")
			return
		}

		// ensure if git is installed
		if !checkIfCommandExists("git") {
			// exit with message
			printError("Git is not installed on your system. Please install git and try again.")
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

		// If not exists, create config file
		mode := cmd.Flag("mode").Value.String()
		var isCreated bool = false
		if mode == "standalone" {
			isCreated = createStandaloneConfig(configFilePath)
		} else if mode == "cluster" {
			isCreated = createClusterConfig(configFilePath)
		} else {
			printError("Invalid mode of operation. Use --mode flag to specify mode of operation")
			return
		}
		if isCreated {
			printSuccess("Config file created at /etc/swiftwave/config.yml")
			printInfo("Run `swiftwave config` to open the config file in editor")
			var msg = ""
			if mode == "standalone" {
				msg =
					`You need to edit at-least the following parameters in the config file:
- Let's Encrypt email address
- Postgres database credentials
`
			} else if mode == "cluster" {
				msg =
					`You need to edit at-least the following parameters in the config file:
- Let's Encrypt email address
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

func createStandaloneConfig(configFilePath string) bool {
	// open and write to file
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			printError(err.Error())
		}
	}(file)
	_, err = file.Write(standaloneConfigSample)
	if err != nil {
		return false
	}
	return true
}

func createClusterConfig(configFilePath string) bool {
	// open and write to file
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			printError(err.Error())
		}
	}(file)
	_, err = file.Write(clusterConfigSample)
	if err != nil {
		return false
	}
	return true
}
