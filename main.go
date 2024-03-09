package main

import (
	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cmd"
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
	// management node also needs docker for running postgres or registry at-least
	_, err = exec.LookPath("docker")
	if err != nil {
		color.Red("Docker is not installed. Aborting.")
		os.Exit(1)
	}
	// Start the command line interface
	cmd.Execute()
}
