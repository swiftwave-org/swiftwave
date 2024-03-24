package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"os"
	"os/exec"
)

var localRegistryContainerName = "swiftwave-image-registry"

func init() {
	localRegistryCmd.AddCommand(startLocalRegistryCmd)
	localRegistryCmd.AddCommand(stopLocalRegistryCmd)
	localRegistryCmd.AddCommand(isLocalRegistryRunningCmd)
	localRegistryCmd.AddCommand(restartLocalRegistryCmd)
}

var localRegistryCmd = &cobra.Command{
	Use:   "localregistry",
	Short: "Manage local image registry for swiftwave service",
	Long:  `Manage local image registry for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var startLocalRegistryCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local image registry",
	Long:  `Start local image registry`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := startLocalRegistry(context.Background()); err != nil {
			printError(err.Error())
		} else {
			printSuccess("Local image registry started")
		}
	},
}

var stopLocalRegistryCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local image registry",
	Long:  `Stop local image registry`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopLocalRegistry(context.Background()); err != nil {
			printError(err.Error())
		} else {
			printSuccess("Local image registry stopped")
		}
	},
}

var isLocalRegistryRunningCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if local image registry is running",
	Long:  `Check if local image registry is running`,
	Run: func(cmd *cobra.Command, args []string) {
		if isRunning, err := isLocalRegistryRunning(context.Background()); err != nil {
			printError(err.Error())
		} else {
			if isRunning {
				printSuccess("Local image registry is running")
			} else {
				printError("Local image registry is not running")
			}
		}
	},
}

var restartLocalRegistryCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart local image registry",
	Long:  `Restart local image registry`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := restartLocalRegistry(context.Background()); err != nil {
			printError(err.Error())
		} else {
			printSuccess("Local image registry restarted")
		}
	},
}

// private functions
func isLocalRegistryRequired() (bool, error) {
	if config.SystemConfig == nil {
		return false, errors.New("system config is not loaded")
	}
	return !config.SystemConfig.ImageRegistryConfig.IsConfigured(), nil
}

func isLocalRegistryRunning(ctx context.Context) (bool, error) {
	if _, err := isLocalRegistryRequired(); err != nil {
		return false, err
	}
	dockerManager, err := containermanger.NewLocalClient(ctx)
	if err != nil {
		return false, err
	}
	return dockerManager.IsContainerRunning(localRegistryContainerName)

}

func startLocalRegistry(ctx context.Context) error {
	isRunning, err := isLocalRegistryRunning(ctx)
	if err != nil {
		return err
	}
	if !isRunning {
		// generate htpasswd file
		htpasswdString, err := config.LocalConfig.LocalImageRegistryConfig.Htpasswd()
		if err != nil {
			return err
		}
		htpasswdString = htpasswdString + "\n"
		// write htpasswd file
		err = os.WriteFile(config.LocalConfig.LocalImageRegistryConfig.AuthPath+"/htpasswd", []byte(htpasswdString), 0611)
		if err != nil {
			return err
		}

		var dockerCmd *exec.Cmd
		if config.LocalConfig.ServiceConfig.UseTLS {
			printInfo("Using TLS for local image registry")
			dockerCmd = exec.Command("docker", "run", "-d",
				"-p", fmt.Sprintf("%d:5000", config.LocalConfig.LocalImageRegistryConfig.Port),
				"--restart", "always",
				"-e", "REGISTRY_AUTH=htpasswd",
				"-e", "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd",
				"-e", "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm",
				"-e", "REGISTRY_HTTP_TLS_CERTIFICATE=/cert/certificate.crt",
				"-e", "REGISTRY_HTTP_TLS_KEY=/cert/private.key",
				"-v", fmt.Sprintf("%s:/cert", config.LocalConfig.LocalImageRegistryConfig.CertPath),
				"-v", fmt.Sprintf("%s:/auth", config.LocalConfig.LocalImageRegistryConfig.AuthPath),
				"-v", fmt.Sprintf("%s:/var/lib/registry", config.LocalConfig.LocalImageRegistryConfig.DataPath),
				"--name", localRegistryContainerName, config.LocalConfig.LocalImageRegistryConfig.Image)
		} else {
			printInfo("Using Non-TLS for local image registry")
			dockerCmd = exec.Command("docker", "run", "-d",
				"-p", fmt.Sprintf("%d:5000", config.LocalConfig.LocalImageRegistryConfig.Port),
				"--restart", "always",
				"-e", "REGISTRY_AUTH=htpasswd",
				"-e", "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd",
				"-e", "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm",
				"-v", fmt.Sprintf("%s:/auth", config.LocalConfig.LocalImageRegistryConfig.AuthPath),
				"-v", fmt.Sprintf("%s:/var/lib/registry", config.LocalConfig.LocalImageRegistryConfig.DataPath),
				"--name", localRegistryContainerName, config.LocalConfig.LocalImageRegistryConfig.Image)
		}
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		err = dockerCmd.Run()
		if err != nil {
			return err
		}
	} else {
		printSuccess("Local image registry is already running")
	}
	return nil
}

func stopLocalRegistry(ctx context.Context) error {
	isRunning, err := isLocalRegistryRunning(ctx)
	if err != nil {
		return err
	}
	if isRunning {
		dockerCmd := exec.Command("docker", "rm", localRegistryContainerName, "--force")
		dockerCmd.Stderr = os.Stderr
		err := dockerCmd.Run()
		if err != nil {
			return err
		}
	} else {
		printSuccess("Local image registry is not running")
	}
	return nil
}

func restartLocalRegistry(ctx context.Context) error {
	err := stopLocalRegistry(ctx)
	if err != nil {
		return err
	}
	return startLocalRegistry(context.Background())
}
