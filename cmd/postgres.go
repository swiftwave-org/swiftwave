package cmd

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/system_config"
)

var postgresContainerName = "swiftwave-postgres"

/*
It will spin up a local postgres container and bind to local machine:
It is not a service and part of swarm, and it's only for standalone installations.

Name : `postgresContainerName`
Image : bitnami/postgresql:latest
Environment Variables:
- POSTGRESQL_DATABASE: <pick_from_config>
- POSTGRESQL_USERNAME: <pick_from_config>
- POSTGRESQL_PASSWORD: <pick_from_config>
- POSTGRESQL_TIMEZONE: <pick_from_config>
Volume Mounts:
- /var/lib/swiftwave/postgres:/bitnami/postgresql
Ports:
- <from_config_ip>:<from_config_port>:5432

Sample Run Command:
docker run -d --name swiftwave-postgres \
	-e POSTGRESQL_DATABASE=swiftwave \
	-e POSTGRESQL_USERNAME=swiftwave \
	-e POSTGRESQL_PASSWORD=swiftwave \
	-e POSTGRESQL_TIMEZONE=Asia/Kolkata \
	-v /var/lib/swiftwave/postgres:/bitnami/postgresql \
	--user 0:0 \
	-p 127.0.0.1:5432:5432 \
	bitnami/postgresql:latest
*/

func init() {
	postgresCmd.AddCommand(postgresStatusCmd)
	postgresCmd.AddCommand(postgresStartCmd)
	postgresCmd.AddCommand(postgresStopCmd)
	rootCmd.AddCommand(postgresCmd)
}

var postgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Manage local postgres database",
	Long:  "Manage local postgres database (Recommended only for standalone installations)",
}

var postgresStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local postgres database",
	Long:  "Start local postgres database (Recommended only for standalone installations)",
	Run: func(cmd *cobra.Command, args []string) {
		// If mode is not standalone, then return
		if systemConfig.Mode != system_config.Standalone {
			printError("Local postgres database is only for standalone installations\n" +
				"For cluster installations, please use a remote postgres database for a better reliability")
			return
		}
		// Create /var/lib/swiftwave/postgres directory if it doesn't exist
		err := createFolder("/var/lib/swiftwave/postgres")
		if err != nil {
			printError("Failed to create folder > /var/lib/swiftwave/postgres")
			return
		}
		// Check if postgres container is already running
		if checkIfPostgresContainerIsRunning() {
			printError("Local postgres database is already running")
			return
		}
		// Check if postgres container exists
		if checkIfPostgresContainerExists() {
			// remove the container
			dockerCmd := exec.Command("docker", "rm", postgresContainerName)
			dockerCmd.Stderr = os.Stderr
			err = dockerCmd.Run()
			if err != nil {
				printError("Failed to remove existing local postgres database")
				return
			}
		}

		// create folder at /var/lib/swiftwave/postgres if not exists
		if !checkIfFolderExists("/var/lib/swiftwave/postgres") {
			err := createFolder("/var/lib/swiftwave/postgres")
			if err != nil {
				printError("Failed to create folder > /var/lib/swiftwave/postgres")
				return
			}
		}

		// Create postgres container
		dockerCmd := exec.Command("docker", "run", "-d", "--name", postgresContainerName,
			"-e", "POSTGRESQL_DATABASE="+systemConfig.PostgresqlConfig.Database,
			"-e", "POSTGRESQL_USERNAME="+systemConfig.PostgresqlConfig.User,
			"-e", "POSTGRESQL_PASSWORD="+systemConfig.PostgresqlConfig.Password,
			"-e", "POSTGRESQL_TIMEZONE="+systemConfig.PostgresqlConfig.TimeZone,
			"-v", "/var/lib/swiftwave/postgres:/bitnami/postgresql:rw",
			"-p", systemConfig.PostgresqlConfig.Host+":"+strconv.Itoa(systemConfig.PostgresqlConfig.Port)+":5432",
			"--user", "0:0",
			"bitnami/postgresql:latest")
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		err = dockerCmd.Run()
		if err != nil {
			printError("Failed to create local postgres database")
			return
		} else {
			printSuccess("Local postgres database started successfully")
		}
	},
}

var postgresStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local postgres database",
	Long:  "Stop local postgres database (Recommended only for standalone installations)",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if postgres container is already running
		if checkIfPostgresContainerIsRunning() {
			// Stop the container
			dockerCmd := exec.Command("docker", "rm", postgresContainerName, "--force")
			dockerCmd.Stderr = os.Stderr
			err := dockerCmd.Run()
			if err != nil {
				printError("Failed to stop local postgres database")
				return
			} else {
				printSuccess("Local postgres database stopped successfully")
			}
		} else {
			printError("Local postgres database is not running")
		}
	},
}

var postgresStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of local postgres database",
	Long:  "Check status of local postgres database (Recommended only for standalone installations)",
	Run: func(cmd *cobra.Command, args []string) {
		if checkIfPostgresContainerIsRunning() {
			printSuccess("Local postgres database is running")
		} else {
			printError("Local postgres database is not running")
		}
	},
}

// Private functions
func checkIfPostgresContainerExists() bool {
	// Use local docker client to check if postgres container exists
	// Check by docker ps -a --format '{{.Names}}' | grep -q "^$container_name$"
	cmd := exec.Command("sh", "-c", "docker ps -a --format '{{.Names}}' | grep -q '^"+postgresContainerName+"$'")
	err := cmd.Run()
	return err == nil
}

func checkIfPostgresContainerIsRunning() bool {
	// Use local docker client to check if postgres container is running
	// Check by docker ps --format '{{.Names}}' | grep -q "^$container_name$"
	cmd := exec.Command("sh", "-c", "docker ps --format '{{.Names}}' | grep -q '^"+postgresContainerName+"$'")
	err := cmd.Run()
	return err == nil
}
