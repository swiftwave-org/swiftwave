package main

import (
	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cmd"
	local_config2 "github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"
	"os/exec"
)

func main() {
	// ensure program is run as root
	if os.Geteuid() != 0 {
		color.Red("This program must be run as root. Aborting.")
		os.Exit(1)
	}
	var err error
	// ensure docker is installed
	_, err = exec.LookPath("docker")
	if err != nil {
		color.Red("Docker is not installed. Aborting.")
		os.Exit(1)
	}
	var config *local_config2.Config
	// Check whether first argument is "install" or no arguments
	if (len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "completion" || os.Args[1] == "--help")) ||
		len(os.Args) == 1 {
		config = nil
	} else {
		// Load the config
		config, err := local_config2.Fetch()
		if err != nil {
			color.Red(err.Error())
			color.Blue("Please run 'swiftwave init' to initialize a configuration file.")
			os.Exit(1)
		}
		// Set the development mode to false
		config.IsDevelopmentMode = false
	}
	// Start the command line interface
	cmd.Execute(config)
}
