package main

import (
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
		color.Red("Docker is not installed. Aborting.")
		os.Exit(1)
	}
	// Start the command line interface
	cmd.Execute()
}
