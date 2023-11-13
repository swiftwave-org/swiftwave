package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/cmd"
	"github.com/swiftwave-org/swiftwave/system_config"
	"os"
)

func main() {
	// ensure program is run as root
	if os.Geteuid() != 0 {
		color.Red("This program must be run as root. Aborting.")
		os.Exit(1)
	}
	var config *system_config.Config
	var err error
	// Check whether first argument is "install" or no arguments
	if (len(os.Args) > 1 && (os.Args[1] == "install" || os.Args[1] == "init" || os.Args[1] == "config" || os.Args[1] == "completion" || os.Args[1] == "--help")) ||
		len(os.Args) == 1 {
		config = nil
	} else {
		// Load config path from environment variable
		systemConfigPath := "/etc/swiftwave/config.yaml"
		// Load the config
		config, err = system_config.ReadFromFile(systemConfigPath)
		if err != nil {
			fmt.Println("failed to load config file > ", err)
			os.Exit(1)
		}
	}

	// Start the command line interface
	cmd.Execute(config)
}
