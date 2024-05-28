package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/moby/sys/user"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cmd"
	"os"
	"os/exec"
	"strings"
)

func main() {
	isNonRoot := false
	// check for ALLOW_NON_ROOT environment variable
	if _, ok := os.LookupEnv("ALLOW_NON_ROOT"); ok {
		if strings.Compare(os.Getenv("ALLOW_NON_ROOT"), "1") == 0 {
			// Prerequisites for using non-root and non-sudo user :
			// Access required for the following:
			// - Docker daemon access (append the user in docker group)
			// - /var/run/swiftwave (need to be created beforehand)
			// - /var/log/swiftwave (need to be created beforehand)
			// - /var/lib/swiftwave (need to be created beforehand)
			// Allow swiftwave binary to access non-root ports
			// sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/bin/swiftwave
			// Need to run haproxy, udp proxy service as non-root
			isNonRoot = true
			username, err := user.CurrentUser()
			if err != nil {
				color.Red("Error getting current user. Aborting.")
				os.Exit(1)
			}
			color.Yellow("[EXPERIMENTAL] Running as non-root user. Please ensure that the user has the required permissions.")
			color.Blue("Running as non-root user.")
			color.Blue("Current user: " + username.Name)
		}
	}

	// ensure program is run as root
	if !isNonRoot && os.Geteuid() != 0 {
		color.Red("This program must be run as root. Aborting.")
		os.Exit(1)
	}
	var err error
	// ensure docker is installed
	// management node also needs docker for running postgres or registry at-least
	_, err = exec.LookPath("docker")
	if err != nil {
		color.Red("Docker is not installed.")
		isDockerInstalled := false
		color.Blue("Run `curl -fsSL get.docker.com | bash -` to install docker.")
		color.Blue("Do you want to install docker now? (y/n)")
		fmt.Print("> ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			color.Red("Error reading response. Aborting.")
			os.Exit(1)
		}
		if strings.Compare(response, "y") == 0 || strings.Compare(response, "Y") == 0 {
			color.Blue("Installing docker...")
			// install docker
			err = runCommand(exec.Command("bash", "-c", "curl -fsSL get.docker.com | bash -"))
			if err != nil {
				color.Red("Error installing docker. Aborting.")
				os.Exit(1)
			}
			// enable docker service
			err = runCommand(exec.Command("systemctl", "enable", "docker"))
			if err != nil {
				color.Red("Error enabling docker. Aborting.")
				os.Exit(1)
			}
			// start docker service
			err = runCommand(exec.Command("systemctl", "start", "docker"))
			if err != nil {
				color.Red("Error starting docker. Aborting.")
				os.Exit(1)
			}
			isDockerInstalled = true
		}
		if !isDockerInstalled {
			color.Red("Docker is not installed. Aborting.")
			os.Exit(1)
		} else {
			color.Green("Docker is installed.")
		}
	}
	// Start the command line interface
	cmd.Execute()
}

// private function
func runCommand(command *exec.Cmd) error {
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}
