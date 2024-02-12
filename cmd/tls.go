package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func init() {
	tlsCmd.AddCommand(tlsEnableCmd)
	tlsCmd.AddCommand(tlsDisableCmd)
	tlsCmd.AddCommand(generateCertificateCommand)
	generateCertificateCommand.Flags().String("domain", "", "Domain name for which to generate the certificate")
}

var tlsCmd = &cobra.Command{
	Use:   "tls",
	Short: "Manage TLS for swiftwave service",
	Long:  `Manage TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var tlsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable TLS for swiftwave service",
	Long:  `Enable TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		if systemConfig.ServiceConfig.UseTLS {
			printSuccess("TLS is already enabled")
			return
		}
		// Check if some certificate is already present
		if isFolderEmpty(systemConfig.ServiceConfig.SSLCertificateDir) {
			printError("No TLS certificate found")
			printInfo("Use `swiftwave tls generate-certificate` to generate a new certificate")
			return
		}
		systemConfig.ServiceConfig.UseTLS = true
		err := systemConfig.WriteToFile(configFilePath)
		if err != nil {
			printError("Failed to update config")
			printError(err.Error())
			return
		}
		printSuccess("TLS has been enabled")
		restartSysctlService("swiftwave")
	},
}

var tlsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable TLS for swiftwave service",
	Long:  `Disable TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		if !systemConfig.ServiceConfig.UseTLS {
			printSuccess("TLS is already disabled")
			return
		}
		systemConfig.ServiceConfig.UseTLS = false
		err := systemConfig.WriteToFile(configFilePath)
		if err != nil {
			printError("Failed to update config")
			printError(err.Error())
			return
		}
		printSuccess("TLS has been disabled")
		restartSysctlService("swiftwave")
	},
}

var generateCertificateCommand = &cobra.Command{
	Use:   "generate-certificate",
	Short: "Generate TLS certificate for swiftwave endpoints",
	Long: `This command generates TLS certificate for swiftwave endpoints.
	It's not for generating certificates for domain of hosted applications`,
	Run: func(cmd *cobra.Command, args []string) {
		// If domain is not provided, use the domain from config
		domain := cmd.Flag("domain").Value.String()
		if strings.TrimSpace(domain) == "" {
			domain = systemConfig.ServiceConfig.AddressOfCurrentNode
		}
		//// Start http-01 challenge server
		echoServer := echo.New()
		echoServer.HideBanner = true
		echoServer.Pre(middleware.RemoveTrailingSlash())
		// Initiating database client
		dbClient, err := getDBClient()
		if err != nil {
			printError("Failed to connect to database")
			return
		}
		// Initiating SSL Manager
		options := SSL.ManagerOptions{
			IsStaging:                 systemConfig.LetsEncryptConfig.StagingEnvironment,
			Email:                     systemConfig.LetsEncryptConfig.EmailID,
			AccountPrivateKeyFilePath: systemConfig.LetsEncryptConfig.AccountPrivateKeyPath,
		}
		sslManager := SSL.Manager{}
		err = sslManager.Init(context.Background(), *dbClient, options)
		if err != nil {
			printError("Failed to initiate SSL Manager")
			return
		}
		// Check if there is already someone listening on port 80
		isPort80Blocked := checkIfPortIsInUse("80")
		isServicePortBlocked := checkIfPortIsInUse(strconv.Itoa(systemConfig.ServiceConfig.BindPort))
		if isPort80Blocked {
			if isServicePortBlocked {
				printInfo("Running swiftwave service will be used to solve http-01 challenge")
			} else {
				printError("Please stop the service running on port 80 temporarily")
				return
			}
		} else {
			// Start the server
			go func(sslManager *SSL.Manager) {
				sslManager.InitHttpHandlers(echoServer)
				err := echoServer.Start(":80")
				if err != nil {
					if errors.Is(err, http.ErrServerClosed) {
						printSuccess("http-01 challenge server has been stopped")
					} else {
						printError("Failed to start http-01 challenge server")
						os.Exit(1)
					}
				}
			}(&sslManager)
		}
		// Generate private key
		privateKey, err := generatePrivateKey()
		if err != nil {
			printError("Failed to generate private key")
			return
		}
		// Generate the certificate
		certificate, err := sslManager.ObtainCertificate(domain, privateKey)
		if err != nil {
			println(err.Error())
			printError("Failed to generate certificate")
			return
		}
		// Stop the http-01 challenge server
		err = echoServer.Server.Shutdown(context.Background())
		if err != nil {
			return
		}
		// Store private key and certificate in the service.ssl_certificate_dir/<domain> folder
		dir := systemConfig.ServiceConfig.SSLCertificateDir + "/" + domain
		if !checkIfFolderExists(dir) {
			err = createFolder(dir)
			if err != nil {
				printError("Failed to create folder " + dir)
				return
			}
		}
		// Store private key
		err = os.WriteFile(dir+"/private.key", []byte(privateKey), 0644)
		if err != nil {
			printError("Failed to store private key")
			return
		}
		// Store certificate
		err = os.WriteFile(dir+"/certificate.crt", []byte(certificate), 0644)
		if err != nil {
			printError("Failed to store certificate")
			return
		}
		// Print success message
		printSuccess("Successfully generated TLS certificate for " + domain)
		// Restart swiftwave service
		restartSysctlService("swiftwave")
	},
}

// private functions

func generatePrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", errors.New("unable to generate private key")
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemKey := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyBytes = pem.EncodeToMemory(&pemKey)
	return string(privateKeyBytes), nil
}

func isFolderEmpty(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return true
	}
	return len(files) == 0
}

func restartSysctlService(serviceName string) {
	// check if service is running
	// read the output of systemctl is-active <service_name>
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return
	}
	if strings.TrimSpace(string(output)) == "active" {
		// restart the service
		cmd = exec.Command("systemctl", "restart", serviceName)
		err = cmd.Run()
		if err != nil {
			return
		}
		printSuccess(serviceName + " has been restarted")
	}
}
