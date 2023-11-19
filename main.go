package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/cmd"
	"github.com/swiftwave-org/swiftwave/system_config"
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
	// ensure docker swarm is initialized
	if !isSwarmInitailized() {
		color.Red("Docker swarm is not initialized. Aborting.")
		color.Blue("Please run 'docker swarm init' to initialize docker swarm node.")
		color.Blue("If you are setting up cluster, you can join the cluster by `docker swarm join` command.")
		os.Exit(1)
	}
	var config *system_config.Config
	// Check whether first argument is "install" or no arguments
	if (len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "completion" || os.Args[1] == "--help")) ||
		len(os.Args) == 1 {
		config = nil
	} else {
		// Load config path from environment variable
		systemConfigPath := "/etc/swiftwave/config.yml"
		// Load the config
		config, err = system_config.ReadFromFile(systemConfigPath)
		if err != nil {
			color.Red(err.Error())
			color.Blue("Please run 'swiftwave init' to initialize a configuration file.")
			os.Exit(1)
		}
	}

	// Start the command line interface
	cmd.Execute(config)
}

// private function
func isSwarmInitailized() bool {
	cmd := exec.Command("docker", "info", "--format", "{{.Swarm.LocalNodeState}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}
